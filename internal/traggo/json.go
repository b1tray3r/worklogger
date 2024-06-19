package traggo

import (
	"strconv"
	"strings"
	"time"
)

type JsonEntry struct {
	Start string `json:"start_utc"`
	End   string `json:"end_utc"`
	Tags  string `json:"tags"`
	Note  string `json:"note"`
}

type Tag struct {
	ID    int64
	Name  string
	Value string
}

type Entry struct {
	Start      time.Time
	End        time.Time
	Duration   time.Duration
	Note       string
	IssueID    int64
	ActionName string
	ActionID   int64
}

func NewEntry(element *JsonEntry) *Entry {
	entry := &Entry{}

	format := "2006-01-02 15:04:05+00:00"
	entry.Start, _ = time.Parse(format, element.Start)
	entry.End, _ = time.Parse(format, element.End)

	entry.Duration = entry.End.Sub(entry.Start)

	entry.Note = element.Note

	tags := strings.Split(element.Tags, " ")
	for _, tag := range tags {
		parts := strings.Split(tag, ":")
		if len(parts) != 2 {
			continue
		}

		parts[0] = strings.TrimPrefix(parts[0], "#")
		switch parts[0] {
		case "issue":
			entry.IssueID, _ = strconv.ParseInt(parts[1], 10, 64)
		case "action":
			entry.ActionName = parts[1]
		}
	}

	return entry
}
