package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func main() {
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
				Name: "add",
				Commands: []*cli.Command{
					{
						Name:  "tag",
						Usage: "add a new tag",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "tag name",
							},
							&cli.StringFlag{
								Name:  "value",
								Usage: "tag value",
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
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

							tagValue := strings.ToLower(cmd.String("value"))
							tag, err := worklogger.AddTag(cmd.String("name"), tagValue)
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
