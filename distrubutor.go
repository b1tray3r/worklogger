package main

import (
	"fmt"
	"sort"
)

// Bucket is a collection of TimeEntry within a specific time frame
type Bucket struct {
	Entries []TimeEntry
}

// TotalHours returns the total hours of all TimeEntry in the Bucket
func (b *Bucket) TotalHours() float64 {
	var total float64

	for _, entry := range b.Entries {
		total += float64(entry.Duration)
	}

	return total
}

// Distributor is a struct that distributes TimeEntry into Buckets
type Distributor struct {
	Buckets       []Bucket
	PauseDuration int
	WorkDuration  int
}

func NewDistributor(amountBuckets, pauseDuration, workDuration int) *Distributor {
	distributor := &Distributor{
		PauseDuration: pauseDuration,
		WorkDuration:  workDuration,
	}

	for i := 0; i < amountBuckets; i++ {
		distributor.Buckets = append(distributor.Buckets, Bucket{})
	}

	return distributor
}

// split will split TimeEntry into smaller TimeEntry with WorkDuration as the maximum hours
func (d *Distributor) split(entries []TimeEntry) []TimeEntry {
	var result []TimeEntry

	for _, entry := range entries {
		for float64(entry.Hours) > float64(d.WorkDuration) {
			result = append(result, TimeEntry{
				ID:         entry.ID,
				Duration:   float64(d.WorkDuration),
				Tags:       entry.Tags,
				Comment:    entry.Comment,
				ActivityID: entry.ActivityID,
				errors:     entry.errors,
				IsRedmine:  entry.IsRedmine,
				IsJira:     entry.IsJira,
			})

			entry.Duration -= float64(d.WorkDuration)
		}

		result = append(result, TimeEntry{
			Duration:   entry.Duration,
			Tags:       entry.Tags,
			Comment:    entry.Comment,
			ActivityID: entry.ActivityID,
			errors:     entry.errors,
			IsRedmine:  entry.IsRedmine,
			IsJira:     entry.IsJira,
		})
	}

	// sort result asc by Hours
	sort.Slice(result, func(i, j int) bool {
		return result[i].Duration > result[j].Duration
	})

	return result
}

func (d *Distributor) Distribute(entries []TimeEntry) []Bucket {
	te := d.split(entries)

	fmt.Println("Total Entries: ", len(te))

	if len(te) == 0 {
		return d.Buckets
	}

	for i, bucket := range d.Buckets {
		for _, entry := range te {
			if bucket.TotalHours()+entry.Duration <= float64(d.WorkDuration) {
				bucket.Entries = append(bucket.Entries, entry)
				d.Buckets[i] = bucket
				te = te[1:]
			}
		}
	}

	return d.Buckets
}
