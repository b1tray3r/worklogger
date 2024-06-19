package redmine

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

type Logger struct {
	APIKey       string
	URL          string
	TicketPrefix string
}

func (rl Logger) getApi() (*redmine.Context, error) {
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

func (rl Logger) Log(issueID int64, activityID int64, startDate time.Time, duration time.Duration, comment string) error {
	api, err := rl.getApi()
	if err != nil {
		return err
	}

	date := startDate.Format("2006-01-02")

	cte, code, err := api.TimeEntryCreate(
		redmine.TimeEntryCreate{
			TimeEntry: redmine.TimeEntryCreateObject{
				IssueID:    &issueID,
				ActivityID: activityID,
				Hours:      duration.Hours(),
				SpentOn:    &date,
				Comments:   comment,
			},
		},
	)
	if err != nil {
		return err
	}
	if code != http.StatusCreated {
		return fmt.Errorf("could not log time entry")
	}

	fmt.Println("logged time entry")
	fmt.Println(cte)
	fmt.Println("Return code:")
	fmt.Println(code)

	return nil
}

func (rl Logger) GetActionID(issueID int64, actionName string) (int64, error) {
	api, err := rl.getApi()
	if err != nil {
		return 0, fmt.Errorf("error getting redmine api: %v", err)
	}

	issue, code, err := api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
	if code == 403 {
		return 0, fmt.Errorf("forbidden to get issue %d", issueID)
	}
	if code != 200 {
		return 0, fmt.Errorf("could not get issue")
	}
	if err != nil {
		return 0, fmt.Errorf("error getting issue: %v", err)
	}

	pID := strconv.FormatInt(issue.Project.ID, 10)
	rP, code, err := api.ProjectSingleGet(
		pID,
		redmine.ProjectSingleGetRequest{
			Includes: []redmine.ProjectInclude{redmine.ProjectIncludeTimeEntryActivities},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("error getting project: %v", err)
	}
	if code != 200 {
		return 0, fmt.Errorf("could not get project")
	}

	if rP.TimeEntryActivities == nil {
		return 0, fmt.Errorf("no time entry activities found for issue %d", issueID)
	}

	resultID := int64(0)
	for _, activity := range *rP.TimeEntryActivities {
		if activity.Name == actionName {
			resultID = activity.ID
			break
		}

		// string begins with actionName
		if len(activity.Name) > len(actionName) && activity.Name[:len(actionName)] == actionName {
			resultID = activity.ID
			break
		}
	}

	return resultID, nil
}
