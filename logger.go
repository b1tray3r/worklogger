package main

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
)

type TimeLogger interface {
	Check(TimeEntry) (bool, error)
	Log(TimeEntry) error
}

type RedmineLogger struct {
	APIKey       string
	URL          string
	TicketPrefix string
}

func (rl RedmineLogger) Check(te TimeEntry) (bool, error) {
	return true, nil
}

func (rl RedmineLogger) Log(te TimeEntry) error {
	return nil
}

type JiraLogger struct {
	Username     string
	Password     string
	URL          string
	TicketPrefix string
}

func (jl JiraLogger) getJiraClient() (*jira.Client, error) {
	tp := jira.BearerAuthTransport{Token: jl.Password}

	client, err := jira.NewClient(jl.URL, tp.Client())
	if err != nil {
		return nil, err
	}

	u, _, err := client.User.GetSelf(context.Background())
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, fmt.Errorf("could not find user - failed to connect maybe")
	}

	return client, nil
}

func (jl JiraLogger) getIssueID(issueIDs []string) (string, error) {
	for _, issueID := range issueIDs {
		if issueID[:len(jl.TicketPrefix)] == jl.TicketPrefix {
			return issueID, nil
		}
	}

	return "", nil
}

func (jl JiraLogger) getIssue(client *jira.Client, issueID string) (*jira.Issue, error) {
	issue, response, err := client.Issue.Get(context.Background(), issueID, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("could not find issue")
	}

	return issue, nil
}

func (jl JiraLogger) Check(te TimeEntry) (bool, error) {
	client, err := jl.getJiraClient()
	if err != nil {
		return false, err
	}

	issueID, err := jl.getIssueID(te.IssueIDs)
	if err != nil {
		return false, err
	}

	issue, err := jl.getIssue(client, issueID)
	if err != nil {
		return false, err
	}

	if issue == nil {
		return false, nil
	}

	return true, nil
}

func (jl JiraLogger) Log(te TimeEntry) error {
	return nil
}
