package main

import (
	"testing"
)

func TestDistributor(t *testing.T) {
	distributor := Distrubutor{
		PauseDuration: 1,
		WorkDuration:  4,
	}

	te := []TimeEntry{
		{
			Hours:      10,
			Tags:       []string{"tag1", "tag2"},
			Comment:    "comment1",
			ActivityID: "activity1",
			errors:     []string{"error1", "error2"},
			IsRedmine:  false,
			IsJira:     false,
		},
		{
			Hours:      5,
			Tags:       []string{"tag3", "tag4"},
			Comment:    "comment2",
			ActivityID: "activity2",
			errors:     []string{"error3", "error4"},
			IsRedmine:  true,
			IsJira:     true,
		},
	}

	result, err := distributor.Distribute(te)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(result) != 5 {
		t.Errorf("Expected 3 entries, got %d", len(result))
	}
}
