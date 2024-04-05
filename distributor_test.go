package main

import (
	"fmt"
	"testing"
)

func TestDistributorOneBucket(t *testing.T) {
	tests := map[string]struct {
		te     []TimeEntry
		amount int
	}{
		"test1": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 1.0,
				},
			},
			amount: 1.0,
		},
		"test2": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 2.0,
				},
				{
					ID:       "2",
					Duration: 1.0,
				},
			},
			amount: 3.0,
		},
		"test3": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 3.0,
				},
				{
					ID:       "2",
					Duration: 1.0,
				},
			},
			amount: 4.0,
		},
		"test4": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 4.0,
				},
			},
			amount: 4.0,
		},
		"test5": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 1.0,
				},
				{
					ID:       "2",
					Duration: 1.0,
				},
				{
					ID:       "3",
					Duration: 1.0,
				},
				{
					ID:       "4",
					Duration: 1.0,
				},
			},
			amount: 4.0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			distributor := NewDistributor(3, 1, 4)

			fmt.Println("OneBucket -------- " + name)
			result := distributor.Distribute(tc.te)

			sum := 0.0
			for i, bucket := range result {
				for _, entry := range bucket.Entries {
					fmt.Printf("Duration: %f\n", float64(entry.Duration))
				}

				if float64(bucket.TotalHours()) > 4 {
					t.Errorf("Bucket %d should not have more than 4 hours", i)
				}

				sum += float64(bucket.TotalHours())
			}

			if sum != float64(tc.amount) {
				t.Errorf("Total hours should be %f, got %f", float64(tc.amount), sum)
			}
		})
	}
}

func TestDistributorTwoBucket(t *testing.T) {
	tests := map[string]struct {
		te     []TimeEntry
		amount int
	}{
		"test1": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 5.0,
				},
			},
			amount: 5.0,
		},
		"test2": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 4.0,
				},
				{
					ID:       "2",
					Duration: 1.0,
				},
			},
			amount: 5.0,
		},
		"test3": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 4.0,
				},
				{
					ID:       "2",
					Duration: 2.0,
				},
			},
			amount: 6.0,
		},
		"test4": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 5.0,
				},
				{
					ID:       "2",
					Duration: 2.0,
				},
			},
			amount: 7.0,
		},
		"test5": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 1.0,
				},
				{
					ID:       "2",
					Duration: 1.0,
				},
				{
					ID:       "3",
					Duration: 1.0,
				},
				{
					ID:       "4",
					Duration: 2.0,
				},
			},
			amount: 5.0,
		},
		"test6": {
			te: []TimeEntry{
				{
					ID:       "1",
					Duration: 3.5,
				},
				{
					ID:       "2",
					Duration: 1.5,
				},
				{
					ID:       "3",
					Duration: 1.5,
				},
				{
					ID:       "4",
					Duration: 0.5,
				},
			},
			amount: 6.0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if name != "test4" {
				t.Skip()
			}

			distributor := NewDistributor(3, 1, 4)

			fmt.Println("TwoBucket -------- " + name)
			result := distributor.Distribute(tc.te)

			sum := 0.0
			for i, bucket := range result {
				fmt.Printf("Bucket: %v\n", i)
				for _, entry := range bucket.Entries {
					fmt.Printf("Duration: %f\n", float64(entry.Duration))
				}

				if float64(bucket.TotalHours()) > 4 {
					t.Errorf("Bucket %d should not have more than 4 hours", i)
				}

				sum += float64(bucket.TotalHours())
			}

			if sum != float64(tc.amount) {
				t.Errorf("Total hours should be %f, got %f", float64(tc.amount), sum)
			}
		})
	}
}
