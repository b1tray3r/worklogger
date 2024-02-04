package main

type TimeLogger interface {
	Check(TimeEntry) (bool, error)
	Log(TimeEntry) error
}

type RedmineLogger struct {
	APIKey string
	URL    string
}

func (rl RedmineLogger) Check(te TimeEntry) (bool, error) {
	return true, nil
}

func (rl RedmineLogger) Log(te TimeEntry) error {
	return nil
}

type JiraLogger struct {
	Username string
	Password string
	URL      string
}

func (jl JiraLogger) Check(te TimeEntry) (bool, error) {
	return true, nil
}

func (jl JiraLogger) Log(te TimeEntry) error {
	return nil
}
