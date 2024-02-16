package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type TimeWarriorEntry struct {
	ID    int
	Start string
	End   string
	Tags  []string
}

type EntryList struct {
	Entries []TimeEntry
}

func (el *EntryList) fromTimeWarrior(time_range string) error {
	timewCommand := fmt.Sprintf("timew export :%s", time_range)
	timewOutput, err := exec.Command("bash", "-c", timewCommand).Output()
	if err != nil {
		return err
	}

	return el.fromJSON(timewOutput)
}

func (el *EntryList) list() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Start", "End", "Hours", "IssueIDs", "Comment"})

	sum := 0.0
	for _, entry := range el.Entries {
		sum += entry.Hours.Hours()
		table.Append([]string{entry.Start.Format("2006-01-02 15:04:05"), entry.End.Format("2006-01-02 15:04:05"), fmt.Sprintf("%.2f", entry.Hours.Hours()), strings.Join(entry.IssueIDs, "\n"), entry.Comment})
	}

	table.SetFooter([]string{" ", "Total", "= " + fmt.Sprintf("%.2f", sum), " ", " "})

	table.Render()
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

		for _, t := range entry.Tags {
			matches := rexp.FindAllString(t, -1)
			for _, match := range matches {
				prefix := match[:2]
				issueID := match[2:]
				switch prefix {
				case "J_":
					tags = append(tags, "PIM-"+issueID)
					isJira = true
				case "R_":
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

func (el *EntryList) fromJSONFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return el.fromJSON(data)
}

type TimeEntry struct {
	IssueIDs  []string
	Start     time.Time
	End       time.Time
	Hours     time.Duration
	Comment   string
	IsRedmine bool
	IsJira    bool
}
