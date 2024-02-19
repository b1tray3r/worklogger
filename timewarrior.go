package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

type TimeWarriorEntry struct {
	ID    int
	Start string
	End   string
	Tags  []string
}

func (el *EntryList) fromJSON(data []byte) error {
	list := make([]TimeWarriorEntry, 0)
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	for _, entry := range list {
		comment := []string{}

		isRedmine := false
		isJira := false

		tags := []string{}
		rexp := regexp.MustCompile(`[RJ]_\w+`)

		// check for existing tag (J or R) in entry.tags
		loggedJ := false
		loggedR := false
		for _, t := range entry.Tags {
			if t == "-J" {
				loggedJ = true
			}
			if t == "-R" {
				loggedR = true
			}
		}

		for _, t := range entry.Tags {
			matches := rexp.FindAllString(t, -1)
			for _, match := range matches {
				prefix := match[:2]
				issueID := match[2:]
				switch prefix {
				case "J_":
					if loggedJ {
						break
					}
					tags = append(tags, "PIM-"+issueID)
					isJira = true
				case "R_":
					if loggedR {
						break
					}
					tags = append(tags, issueID)
					isRedmine = true
				}
				if match == t {
					t = ""
				}
			}

			if t != "" {
				comment = append(comment, t)
			}
		}

		startTime, err := time.Parse("20060102T150405Z", entry.Start)
		if err != nil {
			return err
		}

		if entry.End == "" {
			continue
		}

		endTime, err := time.Parse("20060102T150405Z", entry.End)
		if err != nil {
			return err
		}

		te := TimeEntry{
			IssueIDs:  tags,
			Start:     startTime,
			End:       endTime,
			Hours:     endTime.Sub(startTime),
			Comment:   strings.Join(comment, ", "),
			IsRedmine: isRedmine,
			IsJira:    isJira,
		}
		el.Entries = append(el.Entries, te)
	}

	return nil
}
