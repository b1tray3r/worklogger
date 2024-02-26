package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	redmine "github.com/nixys/nxs-go-redmine/v5"
)

type TimeLogger interface {
	Log(TimeEntry) error
}

type RedmineLogger struct {
	APIKey       string
	URL          string
	TicketPrefix string
}

func (rl RedmineLogger) getApi() (*redmine.Context, error) {
	if rl.URL == "" || rl.APIKey == "" {
		return nil, fmt.Errorf("init error: make sure environment variables `REDMINE_HOST` and `REDMINE_API_KEY` are defined")
	}

	return redmine.Init(
		redmine.Settings{
			Endpoint: rl.URL,
			APIKey:   rl.APIKey,
		},
	), nil
}

func (rl RedmineLogger) getIssueID(issueIDs []string) (int64, error) {
	for _, ID := range issueIDs {
		if ID[:len(rl.TicketPrefix)] == rl.TicketPrefix {
			ID = ID[len(rl.TicketPrefix):]

			issueID, err := strconv.ParseInt(ID, 10, 64)
			if err != nil {
				return 0, err
			}

			return issueID, nil
		}
	}

	return 0, nil
}

func (rl RedmineLogger) Log(te TimeEntry) error {
	ID, err := rl.getIssueID(te.IssueIDs)
	if err != nil {
		return err
	}

	issueID := int64(ID)

	AID := te.ActivityID
	activityID, err := strconv.ParseInt(AID, 10, 64)
	if err != nil {
		return err
	}

	api, err := rl.getApi()
	if err != nil {
		return err
	}
	cte, code, err := api.TimeEntryCreate(
		redmine.TimeEntryCreate{
			TimeEntry: redmine.TimeEntryCreateObject{
				IssueID:    &issueID,
				ActivityID: activityID,
				Hours:      te.Hours.Hours(),
				Comments:   te.Comment,
			},
		},
	)
	if err != nil {
		return err
	}
	if code != http.StatusCreated {
		return fmt.Errorf("could not log time entry")
	}

	fmt.Println(cte)
	fmt.Println(code)

	if code == http.StatusCreated {
		te.mark("S2R")
	}

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

func (jl JiraLogger) Log(te TimeEntry) error {
	client, err := jl.getJiraClient()
	if err != nil {
		return err
	}

	issueID, err := jl.getIssueID(te.IssueIDs)
	if err != nil {
		return err
	}

	issue, err := jl.getIssue(client, issueID)
	if err != nil {
		return err
	}

	var wl struct {
		ID              string `json:"id"`
		StartDate       string `json:"startDate"`
		TimeLogged      string `json:"timeLogged"`
		LogworkCategory string `json:"logworkCategory"`
		Comment         string `json:"comment"`
	}

	wl.ID = issue.ID
	wl.StartDate = te.Start.Format("02/Jan/06 03:04 PM")
	wl.TimeLogged = fmt.Sprintf("%.2f", te.Hours.Hours())
	wl.LogworkCategory = "cat1"
	wl.Comment = te.Comment

	workLog := url.Values{
		"id":              {wl.ID},
		"startDate":       {wl.StartDate},
		"timeLogged":      {wl.TimeLogged},
		"logworkCategory": {wl.LogworkCategory},
		"comment":         {wl.Comment},
	}

	urlStr := jl.URL + "/secure/CreateWorklog.jspa?"

	data := workLog.Encode()
	urlStr += data

	jsonData, err := json.Marshal(wl)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Atlassian-Token", "no-check")
	req.Header.Set("Authorization", "Bearer "+jl.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("could not log work")
	}

	te.mark("S2J")

	return nil
}
