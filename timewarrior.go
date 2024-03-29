package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeWarriorEntry struct {
	ID    int64
	Start string
	End   string
	Tags  []string
}

func parse(entry TimeWarriorEntry) (*TimeEntry, error) {
	startTime, err := time.Parse("20060102T150405Z", entry.Start)
	if err != nil {
		return nil, err
	}

	if entry.End == "" {
		return nil, nil
	}

	endTime, err := time.Parse("20060102T150405Z", entry.End)
	if err != nil {
		return nil, err
	}

	comment := ""
	tags := []string{}

	for _, t := range entry.Tags {
		if strings.Contains(t, " ") {
			comment = t
		} else {
			tags = append(tags, t)
		}
	}

	if comment == "" {
		comment = ""
	} else {
		if strings.Contains(comment, ",") {
			commentParts := strings.Split(comment, ",")
			com := []string{}
			for _, part := range commentParts {
				part = strings.TrimSpace(part)

				if strings.Contains(part, " ") {
					// this is a sentence with spaces, so it's not a tag
					com = append(com, part)
				} else {
					// this is a tag
					tags = append(tags, part)
				}
			}

			comment = strings.Join(com, ", ")
		}
	}

	isJira := false
	isRedmine := false
	rexp := regexp.MustCompile(`[RJA]_\w+`)
	activityID := ""
	issueIDs := []string{}
	tmp := []string{}
	for _, t := range tags {
		matches := rexp.FindAllString(t, -1)
		if len(matches) == 0 {
			tmp = append(tmp, t)
			continue
		}

		for _, match := range matches {
			prefix := match[:2]
			issueID := match[2:]
			switch prefix {
			case "J_":
				isJira = true
				issueIDs = append(issueIDs, "PIM-"+issueID)
			case "R_":
				isRedmine = true
				issueIDs = append(issueIDs, "#"+issueID)
			case "A_":
				activityID = issueID
				tmp = append(tmp, t)
			}
		}
	}

	tags = tmp

	id := strconv.FormatInt(entry.ID, 10)

	return &TimeEntry{
		ID:         id,
		Comment:    comment,
		IssueIDs:   issueIDs,
		Start:      startTime,
		End:        endTime,
		Hours:      endTime.Sub(startTime),
		Tags:       tags,
		IsJira:     isJira,
		IsRedmine:  isRedmine,
		ActivityID: activityID,
	}, nil
}

func (el *EntryList) fromJSON(data []byte) error {
	list := make([]TimeWarriorEntry, 0)
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	for _, entry := range list {
		timeEntry, err := parse(entry)
		if err != nil {
			return err
		}
		if timeEntry == nil {
			continue
		}

		el.Entries = append(el.Entries, *timeEntry)
	}

	return nil
}

func (el *EntryList) getRedmineEntries() []TimeEntry {
	redmineEntries := []TimeEntry{}
	for _, entry := range el.Entries {
		if entry.IsRedmine {
			redmineEntries = append(redmineEntries, entry)
		}
	}
	return redmineEntries
}

func (el *EntryList) getJiraEntries() []TimeEntry {
	jiraEntries := []TimeEntry{}
	for _, entry := range el.Entries {
		if entry.IsJira {
			jiraEntries = append(jiraEntries, entry)
		}
	}
	return jiraEntries
}
