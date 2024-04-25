package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type Log struct {
	State    string
	Logday   string
	Duration float64
	Message  string
	ID       int64
}

type Tag struct {
	ID    int64
	Name  string
	Value string
}

type Config struct {
	DatabaseHost     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
}

type Worklogger struct {
	database *Database
}

func NewWorklogger(config *Config) (*Worklogger, error) {
	db, err := NewDatabase(config.DatabaseUser, config.DatabasePassword, config.DatabaseHost, config.DatabaseName)
	if err != nil {
		fmt.Println("error creating database: ", err)
	}
	return &Worklogger{
		database: db,
	}, nil
}

func (w *Worklogger) Link(logID, tagID int64) error {
	searchQuery := "SELECT * FROM logtags WHERE logid = ? AND tagid = ?;"
	row := w.database.driver.QueryRow(searchQuery, logID, tagID)

	var logtagID int64
	err := row.Scan(&logtagID)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("error linking log and tag: %v", err)
		}
	}

	_, err = w.database.driver.Exec("INSERT INTO logtags (logid, tagid) VALUES (?, ?);", logID, tagID)
	if err != nil {
		return fmt.Errorf("error linking log and tag: %v", err)
	}

	return nil
}

func (w *Worklogger) LoadTag(name, value string) (*Tag, error) {
	if err := w.database.Connect(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	row := w.database.driver.QueryRow("SELECT * FROM tags WHERE name = ? AND value = ?;", name, value)
	tag := &Tag{}
	err := row.Scan(&tag.ID, &tag.Name, &tag.Value)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("error loading tag: %v", err)
		}
	}

	return tag, nil
}

func (w *Worklogger) AddTag(name, value string) (*Tag, error) {
	tag, err := w.LoadTag(name, value)
	if err != nil {
		return nil, fmt.Errorf("error loading tag: %v", err)
	}
	if tag.ID != 0 {
		return tag, nil
	}

	result, err := w.database.driver.Exec("INSERT INTO tags (name, value) VALUES (?, ?);", name, strings.ToLower(value))
	if err != nil {
		return nil, fmt.Errorf("error adding tag: %v", err)
	}

	tagID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting tag ID: %v", err)
	}

	tag = &Tag{
		ID:    tagID,
		Name:  name,
		Value: value,
	}

	return tag, nil
}

func (w *Worklogger) AddLog(message string, tags []string) error {
	if err := w.database.Connect(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer w.database.Close()
	return nil
}
