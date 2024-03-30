package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type TimeEntry struct {
	ID         string
	IssueIDs   []string
	Start      time.Time
	End        time.Time
	Hours      time.Duration
	Tags       []string
	Comment    string
	ActivityID string
	errors     []string
	IsRedmine  bool
	IsJira     bool
}

func (te *TimeEntry) unmark(marker string) error {
	timewCommand := fmt.Sprintf("timew @%s untag %s", te.ID, marker)
	timewOutput, err := exec.Command("bash", "-c", timewCommand).Output()
	if err != nil {
		return err
	}

	log.Printf("%s", timewOutput)
	return nil
}

func (te *TimeEntry) mark(marker string) error {
	timewCommand := fmt.Sprintf("timew @%s tag %s", te.ID, marker)
	timewOutput, err := exec.Command("bash", "-c", timewCommand).Output()
	if err != nil {
		return err
	}

	log.Printf("%s", timewOutput)
	return nil
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

func (el *EntryList) filterPending() {
	var filtered []TimeEntry
	for _, entry := range el.Entries {
		syncFlags := []string{}
		for _, tag := range entry.Tags {
			if tag == "S2R" || tag == "S2J" {
				syncFlags = append(syncFlags, tag)
			}
		}

		if entry.IsRedmine && entry.IsJira && len(syncFlags) < 2 {
			filtered = append(filtered, entry)
		}

		if entry.IsRedmine && len(syncFlags) == 0 {
			filtered = append(filtered, entry)
		}

	}
	el.Entries = filtered
}

func (el *EntryList) list() tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Start", "End", "Hours", "IssueIDs", "Comment", "Tags", "Problems"})

	sum := 0.0
	sum4day := 0.0
	currentDay := ""
	for _, entry := range el.Entries {
		if currentDay == "" {
			currentDay = entry.Start.Format("2006-01-02")
		}

		if currentDay != entry.Start.Format("2006-01-02") {
			table.Append([]string{" ", " ", currentDay, "= " + fmt.Sprintf("%.2f", sum4day), " ", " ", " ", " "})
			sum4day = 0.0
			currentDay = entry.Start.Format("2006-01-02")
		}

		// checking for problems
		if len(entry.Comment) == 0 {
			entry.errors = append(entry.errors, "Comment is empty")
		}

		if entry.IsJira && !entry.IsRedmine {
			entry.errors = append(entry.errors, "Jira entry without Redmine issue")
		}

		if entry.IsRedmine && entry.IsJira && len(entry.IssueIDs) < 2 {
			entry.errors = append(entry.errors, "Redmine and Jira issue without both issue IDs")
		}

		if entry.IsRedmine && entry.ActivityID == "" {
			entry.errors = append(entry.errors, "Redmine entry without activity ID")
		}

		loc, _ := time.LoadLocation("Europe/Berlin")

		sum4day += entry.Hours.Hours()
		sum += entry.Hours.Hours()
		table.Append([]string{
			entry.ID,
			entry.Start.In(loc).Format("2006-01-02 15:04:05"),
			entry.End.In(loc).Format("2006-01-02 15:04:05"),
			fmt.Sprintf(
				"%.2f",
				entry.Hours.Hours(),
			),
			strings.Join(
				entry.IssueIDs,
				"\n",
			),
			entry.Comment,
			strings.Join(
				entry.Tags,
				"\n",
			),
			strings.Join(
				entry.errors,
				"\n",
			),
		})
	}

	table.SetFooter([]string{" ", " ", "Total", "= " + fmt.Sprintf("%.2f", sum), " ", " ", " ", " "})

	return *table
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
