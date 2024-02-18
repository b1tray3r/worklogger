package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var (
	el EntryList
	jl JiraLogger
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
						Name:  "range",
						Value: "month",
						Usage: "The time range to list. Valid ranges are 'month', 'week', and 'day'.",
					},
				},
				Action: func(cCtx *cli.Context) error {
					time_range := cCtx.String("range")
					if time_range != "month" && time_range != "week" && time_range != "day" {
						fmt.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
						return nil
					}

					if err := el.fromTimeWarrior(time_range); err != nil {
						return err
					}
					el.list()

					return nil
				},
			},
			{
				Name:  "log",
				Usage: "Get the time entries from timewarrior and log them to other systems.",
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
					&cli.BoolFlag{
						Name: "not-redmine",
					},
					&cli.BoolFlag{
						Name: "not-jira",
					},
				},
				Action: func(ctx *cli.Context) error {
					log.Printf("%v", ctx.String("jira-url"))
					rl := RedmineLogger{
						APIKey:       ctx.String("redmine-api-token"),
						URL:          ctx.String("redmine-url"),
						TicketPrefix: "",
					}
					jl := JiraLogger{
						Username:     ctx.String("jira-username"),
						Password:     ctx.String("jira-api-token"),
						URL:          ctx.String("jira-url"),
						TicketPrefix: "PIM-",
					}

					time_range := ctx.String("range")
					if time_range != "month" && time_range != "week" && time_range != "day" {
						fmt.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
					}

					if err := el.fromTimeWarrior(time_range); err != nil {
						return err
					}

					for _, entry := range el.Entries {
						// log.Printf("Logging entry: %s", entry.Comment)

						if entry.IsRedmine && !ctx.Bool("not-redmine") {
							success, err := rl.Check(entry)
							if err != nil {
								return err
							}
							if success {
								log.Println("to Redmine:")
								rl.Log(entry)
							}
						}

						if entry.IsJira && !ctx.Bool("not-jira") {
							success, err := jl.Check(entry)
							if err != nil {
								return err
							}

							if success {
								log.Println("\nto Jira:\n")
								jl.Log(entry)
							}
						}
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
