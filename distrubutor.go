package main

import (
	"sort"
	"time"
)

// Bucket is a collection of TimeEntry within a specific time frame
type Bucket struct {
	Entries []TimeEntry
}

// TotalHours returns the total hours of all TimeEntry in the Bucket
func (b *Bucket) TotalHours() float64 {
	var total float64

	for _, entry := range b.Entries {
		total += float64(entry.Hours)
	}

	return total
}

// Distrubutor is a struct that distributes TimeEntry into Buckets
type Distrubutor struct {
	Buckets       []Bucket
	PauseDuration int
	WorkDuration  int
}

// split will split TimeEntry into smaller TimeEntry with WorkDuration as the maximum hours
func (d *Distrubutor) split(entries []TimeEntry) []TimeEntry {
	var result []TimeEntry

	for _, entry := range entries {
		for float64(entry.Hours) > float64(d.WorkDuration) {
			result = append(result, TimeEntry{
				Hours:      time.Duration(d.WorkDuration),
				Tags:       entry.Tags,
				Comment:    entry.Comment,
				ActivityID: entry.ActivityID,
				errors:     entry.errors,
				IsRedmine:  entry.IsRedmine,
				IsJira:     entry.IsJira,
			})

			entry.Hours -= time.Duration(d.WorkDuration)
		}

		result = append(result, TimeEntry{
			Hours:      entry.Hours,
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
		return result[i].Hours > result[j].Hours
	})

	return result
}

func (d *Distrubutor) Distribute(entries []TimeEntry) []Bucket {
	te := d.split(entries)

	var currentBucket Bucket
	for i, entry := range te {
		if float64(entry.Hours) == float64(d.WorkDuration) {
			currentBucket.Entries = append(currentBucket.Entries, entry)

			d.Buckets = append(d.Buckets, currentBucket)
			currentBucket = Bucket{}
			continue
		}

		if currentBucket.TotalHours()+float64(entry.Hours) <= float64(d.WorkDuration) {
			currentBucket.Entries = append(currentBucket.Entries, entry)

			if i+1 == len(te) {
				d.Buckets = append(d.Buckets, currentBucket)
				continue
			}
		}

		if currentBucket.TotalHours() == float64(d.WorkDuration) {
			d.Buckets = append(d.Buckets, currentBucket)
			currentBucket = Bucket{}
		}
	}

	return d.Buckets
}
