package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

var (
	el EntryList
)

func main() {
	el := EntryList{}
	rl := RedmineLogger{}
	jl := JiraLogger{}

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
				Action: func(cCtx *cli.Context) error {
					time_range := cCtx.String("range")
					if time_range != "month" && time_range != "week" && time_range != "day" {
						fmt.Println("Invalid time range. Please use 'month', 'week', or 'day'.")
					}

					if err := el.fromTimeWarrior(time_range); err != nil {
						return err
					}

					for _, entry := range el.Entries {
						log.Printf("Logging entry: %s", entry.Comment)

						if entry.IsRedmine && !cCtx.Bool("not-redmine") {
							success, err := rl.Check(entry)
							if err != nil {
								return err
							}
							if success {
								log.Println("to Redmine:")
								rl.Log(entry)
							}
						}

						if entry.IsJira && !cCtx.Bool("not-jira") {
							success, err := jl.Check(entry)
							if err != nil {
								return err
							}

							if success {
								log.Println("to Jira:")
								jl.Log(entry)
							}
						}

						fmt.Println(entry)
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
