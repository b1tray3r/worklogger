package main

import (
	"fmt"
	"testing"
)

func TestDistributor(t *testing.T) {
	distributor := Distrubutor{
		PauseDuration: 1,
		WorkDuration:  4.0,
	}

	te := []TimeEntry{
		{
			ID:         "1",
			Hours:      10,
			Tags:       []string{"tag1", "tag2"},
			Comment:    "comment1",
			ActivityID: "activity1",
			errors:     []string{"error1", "error2"},
			IsRedmine:  false,
			IsJira:     false,
		},
		{
			ID:         "2",
			Hours:      5,
			Tags:       []string{"tag3", "tag4"},
			Comment:    "comment2",
			ActivityID: "activity2",
			errors:     []string{"error3", "error4"},
			IsRedmine:  true,
			IsJira:     true,
		},
	}

	result := distributor.Distribute(te)

	for i, bucket := range result {
		fmt.Printf("Bucket %d \n", i)
		for _, entry := range bucket.Entries {
			fmt.Printf("\t %s - %d\n", entry.Comment, entry.Hours)
		}
	}

	want := 4
	if len(result) != want {
		t.Errorf("Expected %d buckets, got %d", want, len(result))
	}
}
