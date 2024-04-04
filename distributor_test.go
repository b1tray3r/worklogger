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
					ID:    "1",
					Hours: 1,
				},
			},
			amount: 1,
		},
		"test2": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 2,
				},
				{
					ID:    "2",
					Hours: 1,
				},
			},
			amount: 1,
		},
		"test3": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 3,
				},
				{
					ID:    "2",
					Hours: 1,
				},
			},
			amount: 1,
		},
		"test4": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 4,
				},
			},
			amount: 1,
		},
		"test5": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 1,
				},
				{
					ID:    "2",
					Hours: 1,
				},
				{
					ID:    "3",
					Hours: 1,
				},
				{
					ID:    "4",
					Hours: 1,
				},
			},
			amount: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			distributor := Distrubutor{
				PauseDuration: 1,
				WorkDuration:  4.0,
			}

			fmt.Println("-------- " + name)
			result := distributor.Distribute(tc.te)

			for i, bucket := range result {
				fmt.Printf("Bucket %d\n", i)
				for _, entry := range bucket.Entries {
					fmt.Printf("ID: %s, Hours: %f\n", entry.ID, float64(entry.Hours))
				}
			}

			if len(result) != tc.amount {
				t.Errorf("Expected %d buckets, got %d", tc.amount, len(result))
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
					ID:    "1",
					Hours: 5,
				},
			},
			amount: 2,
		},
		"test2": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 4,
				},
				{
					ID:    "2",
					Hours: 1,
				},
			},
			amount: 2,
		},
		"test3": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 4,
				},
				{
					ID:    "2",
					Hours: 2,
				},
			},
			amount: 2,
		},
		"test4": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 5,
				},
				{
					ID:    "2",
					Hours: 2,
				},
			},
			amount: 2,
		},
		"test5": {
			te: []TimeEntry{
				{
					ID:    "1",
					Hours: 1,
				},
				{
					ID:    "2",
					Hours: 1,
				},
				{
					ID:    "3",
					Hours: 1,
				},
				{
					ID:    "4",
					Hours: 2,
				},
			},
			amount: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			distributor := Distrubutor{
				PauseDuration: 1,
				WorkDuration:  4.0,
			}

			fmt.Println("-------- " + name)
			result := distributor.Distribute(tc.te)

			for i, bucket := range result {
				fmt.Printf("Bucket %d\n", i)
				for _, entry := range bucket.Entries {
					fmt.Printf("ID: %s, Hours: %f\n", entry.ID, float64(entry.Hours))
				}
			}

			if len(result) != tc.amount {
				t.Errorf("Expected %d buckets, got %d", tc.amount, len(result))
			}
		})
	}
}
