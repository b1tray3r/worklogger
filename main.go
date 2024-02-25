package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	err := godotenv.Load()
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
				},
				Action: func(ctx *cli.Context) error {
					rl := &RedmineLogger{
						APIKey:       ctx.String("redmine-api-token"),
						URL:          ctx.String("redmine-url"),
						TicketPrefix: "#",
					}
					_, err := rl.init()
					if err != nil {
						return err
					}

					time_range := ctx.String("range")
					if time_range != "all" && time_range != "month" && time_range != "week" && time_range != "day" {
						log.Println("Invalid time range. Please use 'all', 'month', 'week', or 'day'.")
						return nil
					}

					if err := el.fromTimeWarrior(time_range); err != nil {
						return err
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

					for _, entry := range el.Entries {
						for _, issueID := range entry.IssueIDs {
							if issueID == ctx.String("issueID") {
								entry.unmark(ctx.String("tag"))
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
							rc, err := rl.init()
							if err != nil {
								return err
							}

							time_range := ctx.String("range")
							if time_range != "month" && time_range != "week" && time_range != "day" {
								fmt.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
							}

							if err := el.fromTimeWarrior(time_range); err != nil {
								return err
							}

							projects := &Projects{}
							redmineEntries := []TimeEntry{}
							for _, entry := range el.Entries {
								if entry.IsRedmine {
									redmineEntries = append(redmineEntries, entry)
									issueID, err := rl.getIssueID(entry.IssueIDs)
									if err != nil {
										return err
									}

									issue, code, err := rc.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
									iID := strconv.FormatInt(issue.ID, 10)
									if err != nil {
										log.Printf("Error getting issue %s: %s", iID, err)
										continue
									}
									if code != 200 {
										log.Printf("Error getting issue %s: %d", iID, code)
										continue
									}

									pID := strconv.FormatInt(issue.Project.ID, 10)
									rP, code, err := rc.ProjectSingleGet(
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

									for _, activity := range *rP.TimeEntryActivities {
										project.Activities = append(project.Activities, Activity{
											ID:  strconv.FormatInt(activity.ID, 10),
											Tag: activity.Name,
										})
									}
								}
							}

							log.Printf("Found %d Redmine entries", len(redmineEntries))
							log.Printf("with %d Projects", len(projects.Projects))

							for _, entry := range redmineEntries {
								err := rl.Check(entry)
								if err != nil {
									return err
								}

								issueID, err := rl.getIssueID(entry.IssueIDs)
								if err != nil {
									return err
								}
								iID := strconv.FormatInt(issueID, 10)
								log.Printf("Logging time entry to Redmine: %s", iID)

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

								jl.Log(entry)
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
								success, err := jl.Check(entry)
								if err != nil {
									return err
								}

								if success {
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
