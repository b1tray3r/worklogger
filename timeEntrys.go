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
	ID        string
	IssueIDs  []string
	Start     time.Time
	End       time.Time
	Hours     time.Duration
	Comment   string
	IsRedmine bool
	IsJira    bool
}

func (te *TimeEntry) markSynced(marker string) error {
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

func (el *EntryList) list() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Start", "End", "Hours", "IssueIDs", "Comment"})

	sum := 0.0
	for _, entry := range el.Entries {
		sum += entry.Hours.Hours()
		table.Append([]string{entry.ID, entry.Start.Format("2006-01-02 15:04:05"), entry.End.Format("2006-01-02 15:04:05"), fmt.Sprintf("%.2f", entry.Hours.Hours()), strings.Join(entry.IssueIDs, "\n"), entry.Comment})
	}

	table.SetFooter([]string{" ", "Total", "= " + fmt.Sprintf("%.2f", sum), " ", " "})

	table.Render()
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
