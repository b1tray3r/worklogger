package main

type Distrubutor struct {
	PauseDuration int
	WorkDuration  int
}

func (d *Distrubutor) Distribute(entries []TimeEntry) ([]TimeEntry, error) {
	var result []TimeEntry
	for _, entry := range entries {
		for entry.Hours > 4 {
			result = append(result, TimeEntry{
				Hours:      4,
				Tags:       entry.Tags,
				Comment:    entry.Comment,
				ActivityID: entry.ActivityID,
				errors:     entry.errors,
				IsRedmine:  entry.IsRedmine,
				IsJira:     entry.IsJira,
			})

			entry.Hours -= 4
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
	return result, nil
}
