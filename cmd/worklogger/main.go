package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/b1tray3r/worklogger/internal/redmine"
	"github.com/b1tray3r/worklogger/internal/traggo"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v3"
)

func main() {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Fatal(err)
	}
	time.Local = loc

	cmd := &cli.Command{
		Name:  "worklogger",
		Usage: "log your work",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dbHost",
				Usage: "Database host address",
			},
			&cli.StringFlag{
				Name:  "dbUser",
				Usage: "Database username",
			},
			&cli.StringFlag{
				Name:  "dbPassword",
				Usage: "Database password",
			},
			&cli.StringFlag{
				Name:  "dbName",
				Usage: "Database name",
				Value: "worklogger",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "import",
				Commands: []*cli.Command{
					{
						Name:  "json",
						Usage: "import logs from a JSON file",
						Action: func(_ context.Context, cmd *cli.Command) error {
							fmt.Println("importing logs from JSON file")

							file := cmd.Args().Get(0)
							if file == "" {
								return fmt.Errorf("file is required")
							}

							jsonFile, err := os.Open(file)
							if err != nil {
								fmt.Println(err)
							}
							defer jsonFile.Close()

							decoder := json.NewDecoder(jsonFile)
							jsonEntries := []*traggo.JsonEntry{}
							if err := decoder.Decode(&jsonEntries); err != nil {
								fmt.Println(err)
							}

							entries := []*traggo.Entry{}
							for _, jsonEntry := range jsonEntries {
								entry := traggo.NewEntry(jsonEntry)
								entries = append(entries, entry)
							}

							rl := &redmine.Logger{
								APIKey: os.Getenv("REDMINE_API_KEY"),
								URL:    os.Getenv("REDMINE_HOST"),
							}

							fmt.Printf("importing %d entries\n", len(entries))
							fmt.Printf("%#v\n", rl)

							for _, entry := range entries {
								actionID, err := rl.GetActionID(entry.IssueID, entry.ActionName)
								if err != nil {
									fmt.Printf("error getting action ID: %v - %+v\n", err, entry)
								}

								if actionID == 0 {
									fmt.Printf("action not found: %+v\n", entry)
									continue
								}
								if err := rl.Log(entry.IssueID, actionID, entry.Start, entry.Duration, entry.Note); err != nil {
									fmt.Printf("error logging entry: %v\n", err)
								}
							}

							return nil
						},
					},
				},
			},
			{
				Name: "list",
				Commands: []*cli.Command{
					{
						Name:  "tags",
						Usage: "list all tags",
						Action: func(_ context.Context, cmd *cli.Command) error {
							config := &Config{
								DatabaseHost:     cmd.String("dbHost"),
								DatabaseUser:     cmd.String("dbUser"),
								DatabasePassword: cmd.String("dbPassword"),
								DatabaseName:     cmd.String("dbName"),
							}

							worklogger, err := NewWorklogger(config)
							if err != nil {
								return fmt.Errorf("error creating worklogger: %v", err)
							}

							tags, err := worklogger.ListTags()
							if err != nil {
								return fmt.Errorf("error listing tags: %v", err)
							}

							rows := []table.Row{}
							for _, tag := range tags {
								rows = append(rows, table.Row{tag.ID, tag.Name, tag.Value})
							}

							t := table.NewWriter()
							t.SetOutputMirror(os.Stdout)
							t.AppendHeader(table.Row{"#", "Name", "Value"})
							t.AppendRows(
								rows,
							)
							t.AppendSeparator()
							t.Render()

							return nil
						},
					},
				},
			},
			{
				Name: "add",
				Commands: []*cli.Command{
					{
						Name:  "tag",
						Usage: "add a new tag",
						Action: func(_ context.Context, cmd *cli.Command) error {
							config := &Config{
								DatabaseHost:     cmd.String("dbHost"),
								DatabaseUser:     cmd.String("dbUser"),
								DatabasePassword: cmd.String("dbPassword"),
								DatabaseName:     cmd.String("dbName"),
							}

							name := cmd.Args().Get(0)
							if name == "" {
								return fmt.Errorf("tag name is required")
							}

							value := cmd.Args().Get(1)
							if value == "" {
								return fmt.Errorf("tag value is required")
							}

							worklogger, err := NewWorklogger(config)
							if err != nil {
								return fmt.Errorf("error creating worklogger: %v", err)
							}

							tagName := strings.ToLower(name)
							tagValue := strings.ToLower(value)
							tag, err := worklogger.AddTag(tagName, tagValue)
							if err != nil {
								return fmt.Errorf("error adding tag: %v", err)
							}

							fmt.Printf("tag added: %v\n", tag)
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
