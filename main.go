package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"
	redmine "github.com/nixys/nxs-go-redmine/v5"
	"github.com/urfave/cli"
)

var (
	el EntryList
	jl JiraLogger
	rl RedmineLogger
)

func main() {
	configFile, err := xdg.ConfigFile("worklogger/config.env")
	if err != nil {
		log.Fatal("Error getting config file")
	}
	err = godotenv.Load(configFile)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	el := EntryList{}

	app := &cli.App{
		Name:  "worklogger",
		Usage: "A work logger which can log time to Redmine and JIRA.",
		Flags: []cli.Flag{},
		Commands: []cli.Command{
			{
				Name:  "list",
				Usage: "List the time entries from timewarrior.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "redmine-api-token",
						Usage: "The API key for Redmine.",
						Value: os.Getenv("WL_REDMINE_API_TOKEN"),
					},
					&cli.StringFlag{
						Name:  "redmine-url",
						Usage: "The URL for Redmine.",
						Value: os.Getenv("WL_REDMINE_URL"),
					},
					&cli.StringFlag{
						Name:  "range",
						Value: "all",
						Usage: "The time range to list. Valid ranges are 'all', 'month', 'week', and 'day'.",
					},
					&cli.BoolFlag{
						Name:  "pending",
						Usage: "Show time entries which are not yet synced to Redmine or JIRA.",
					},
				},
				Action: func(ctx *cli.Context) error {
					time_range := ctx.String("range")
					if time_range != "all" && time_range != "month" && time_range != "week" && time_range != "day" {
						log.Println("Invalid time range. Please use 'all', 'month', 'week', or 'day'.")
						return nil
					}

					if err := el.fromTimeWarrior(time_range); err != nil {
						return err
					}

					if ctx.Bool("pending") {
						el.filterPending()
					}

					table := el.list()
					table.Render()

					return nil
				},
			},
			{
				Name:  "untag",
				Usage: "Remove a tag from a list of entries",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tag",
						Usage: "The tag to set to the entries",
					},
					&cli.StringFlag{
						Name:  "issueID",
						Usage: "The issueIDs of the entries to tag",
					},
				},
				Action: func(ctx *cli.Context) error {
					if err := el.fromTimeWarrior("all"); err != nil {
						return err
					}

					tag := ctx.String("tag")
					ID := ctx.String("issueID")

					for _, entry := range el.Entries {
						if ID == "*" {
							entry.unmark(tag)
							continue
						}

						for _, issueID := range entry.IssueIDs {
							if issueID == ID {
								entry.unmark(tag)
							}
						}
					}

					return nil
				},
			},

			{
				Name:  "tag",
				Usage: "Set a tag to a list of entries",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tag",
						Usage: "The tag to set to the entries",
					},
					&cli.StringFlag{
						Name:  "issueID",
						Usage: "The issueIDs of the entries to tag",
					},
				},
				Action: func(ctx *cli.Context) error {
					if err := el.fromTimeWarrior("all"); err != nil {
						return err
					}

					for _, entry := range el.Entries {
						for _, issueID := range entry.IssueIDs {
							if issueID == ctx.String("issueID") {
								entry.mark(ctx.String("tag"))
							}
						}
					}

					return nil
				},
			},
			{
				Name:  "log",
				Usage: "Get the time entries from timewarrior and log them to other systems.",

				Subcommands: []cli.Command{
					{
						Name:  "redmine",
						Usage: "Log the time entries to Redmine.",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "redmine-api-token",
								Usage: "The API key for Redmine.",
								Value: os.Getenv("WL_REDMINE_API_TOKEN"),
							},
							&cli.StringFlag{
								Name:  "redmine-url",
								Usage: "The URL for Redmine.",
								Value: os.Getenv("WL_REDMINE_URL"),
							},
							&cli.StringFlag{
								Name:  "range",
								Value: "month",
								Usage: "The time range to list. Valid ranges are 'month', 'week', and 'day'.",
							},
						},
						Action: func(ctx *cli.Context) error {
							rl := &RedmineLogger{
								APIKey:       ctx.String("redmine-api-token"),
								URL:          ctx.String("redmine-url"),
								TicketPrefix: "#",
							}
							api, err := rl.getApi()
							if err != nil {
								return err
							}

							user, code, err := api.UserCurrentGet(redmine.UserCurrentGetRequest{})
							if code != 200 {
								return fmt.Errorf("error getting user: %d", code)
							}
							if err != nil {
								return fmt.Errorf("error getting user: %s", err)
							}

							log.Printf("Logged in as %s", user.Login)

							time_range := ctx.String("range")
							if time_range != "month" && time_range != "week" && time_range != "day" {
								fmt.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
							}

							if err := el.fromTimeWarrior(time_range); err != nil {
								return err
							}

							projects := &Projects{}
							redmineEntries := []TimeEntry{}
							mapping := map[string]string{}
							for _, entry := range el.Entries {
								if entry.IsRedmine {
									issueID, err := rl.getIssueID(entry.IssueIDs)
									if err != nil {
										return err
									}

									iID := strconv.FormatInt(issueID, 10)
									log.Printf("Checking %s against Redmine.", iID)

									issue, code, err := api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
									if code == 403 {
										entry.errors = append(entry.errors, fmt.Sprintf("Access forbidden on %s: %d", iID, code))
										log.Printf("Access forbidden on %s: %d", iID, code)
										continue
									}
									if code != 200 {
										entry.errors = append(entry.errors, fmt.Sprintf("Unexpected code on %s: %d", iID, code))
										log.Printf("Unexpected code on %s: %d", iID, code)
										continue
									}
									if err != nil {
										entry.errors = append(entry.errors, fmt.Sprintf("Error getting issue %s: %s", iID, err))
										log.Printf("Error getting issue %s: %s", iID, err)
										continue
									}

									// handle activities
									if entry.ActivityID == "" {
										if val, ok := mapping[iID]; ok {
											entry.ActivityID = val
										} else {

											log.Print("No activity ID found, asking the API for the activities...")
											log.Printf("IssueID %s, Comment: %s", iID, entry.Comment)
											pID := strconv.FormatInt(issue.Project.ID, 10)
											rP, code, err := api.ProjectSingleGet(
												pID,
												redmine.ProjectSingleGetRequest{
													Includes: []redmine.ProjectInclude{redmine.ProjectIncludeTimeEntryActivities},
												},
											)
											if err != nil {
												log.Printf("Error getting project %s: %s", issue.Project.Name, err)
												continue
											}
											if code != 200 {
												log.Printf("Error getting project %s: %d", issue.Project.Name, code)
												continue
											}

											project := projects.GetProject(pID)
											if project == nil {
												project = &Project{
													ID:         pID,
													TimeEntrys: []TimeEntry{},
													Activities: []Activity{},
												}
												project.TimeEntrys = append(project.TimeEntrys, entry)
												projects.AddProject(*project)
											}

											if len(project.Activities) <= 0 {
												for _, activity := range *rP.TimeEntryActivities {
													project.Activities = append(project.Activities, Activity{
														ID:  strconv.FormatInt(activity.ID, 10),
														Tag: activity.Name,
													})
												}
											}

											log.Println("=====================================")
											for index, activity := range project.Activities {
												fmt.Printf("%d:\t%s\n", index, activity.Tag)
											}
											log.Println("=====================================")
											var input string
											fmt.Printf("Please enter the number of your activity: ")
											fmt.Scanf("%s", &input)
											selected, err := strconv.Atoi(input)
											if err != nil {
												return err
											}

											entry.ActivityID = project.Activities[selected].ID
											mapping[iID] = entry.ActivityID
										}

										entry.mark(fmt.Sprintf("A_%s", entry.ActivityID))
									}

									redmineEntries = append(redmineEntries, entry)
								}
							}

							for _, entry := range redmineEntries {
								issueID, err := rl.getIssueID(entry.IssueIDs)
								if err != nil {
									return err
								}
								iID := strconv.FormatInt(issueID, 10)
								log.Printf("Logging %s", iID)

								alreadySynced := false
								for _, tag := range entry.Tags {
									if tag == "S2R" {
										alreadySynced = true
										break
									}
								}
								if alreadySynced {
									log.Println(">\tAlready synced to Redmine")
									continue
								}

								if len(entry.errors) > 0 {
									log.Println(">\tSkipping due to errors")
									continue
								}

								err = rl.Log(entry)
								if err != nil {
									return err
								}
							}

							return nil
						},
					},
					{
						Name:  "jira",
						Usage: "Log the time entries to JIRA.",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "jira-username",
								Usage: "The username for JIRA.",
								Value: os.Getenv("WL_JIRA_USERNAME"),
							},
							&cli.StringFlag{
								Name:  "jira-api-token",
								Usage: "The API token for JIRA.",
								Value: os.Getenv("WL_JIRA_API_TOKEN"),
							},
							&cli.StringFlag{
								Name:  "jira-url",
								Usage: "The URL for JIRA.",
								Value: os.Getenv("WL_JIRA_URL"),
							},
							&cli.StringFlag{
								Name:  "range",
								Value: "month",
								Usage: "The time range to list. Valid ranges are 'month', 'week', and 'day'.",
							},
						},
						Action: func(ctx *cli.Context) error {
							jl := JiraLogger{
								Username:     ctx.String("jira-username"),
								Password:     ctx.String("jira-api-token"),
								URL:          ctx.String("jira-url"),
								TicketPrefix: "PIM-",
							}
							time_range := ctx.String("range")
							if time_range != "month" && time_range != "week" && time_range != "day" {
								log.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
							}

							if err := el.fromTimeWarrior(time_range); err != nil {
								return err
							}

							jiraEntries := []TimeEntry{}
							for _, entry := range el.Entries {
								if entry.IsJira {
									jiraEntries = append(jiraEntries, entry)
								}
							}

							log.Printf("Found %d JIRA entries", len(jiraEntries))

							for _, entry := range jiraEntries {
								issueID, err := jl.getIssueID(entry.IssueIDs)
								if err != nil {
									return err
								}
								log.Printf("Logging time entry to JIRA: %s", issueID)

								alreadySynced := false
								for _, tag := range entry.Tags {
									if tag == "S2J" {
										alreadySynced = true
										break
									}
								}
								if alreadySynced {
									log.Println(">\tAlready synced to JIRA")
									continue
								}

								jl.Log(entry)
							}

							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
