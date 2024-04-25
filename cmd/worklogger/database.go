package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var ErrNotFound = errors.New("entry not found")

type Database struct {
	driver   *sql.DB
	username string
	password string
	hostname string
	dbname   string
}

func NewDatabase(username, password, hostname, dbname string) (*Database, error) {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+hostname+")/"+dbname)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	tables := []string{"1-table-tags.sql", "2-table-logs.sql", "3-table-logtags.sql"}
	for _, filename := range tables {
		createStatement, err := os.ReadFile("files/database/" + filename)
		if err != nil {
			return nil, fmt.Errorf("error reading table creation file: %v", err)
		}

		_, err = db.Exec(string(createStatement))
		if err != nil {
			return nil, fmt.Errorf("error creating logs table: %v", err)
		}
	}

	return &Database{
		username: username,
		password: password,
		hostname: hostname,
		dbname:   dbname,
	}, nil
}

func (d *Database) dsn() string {
	return d.username + ":" + d.password + "@tcp(" + d.hostname + ")/" + d.dbname
}

func (d *Database) Connect() error {
	db, err := sql.Open("mysql", d.dsn())
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	d.driver = db
	return nil
}

func (d *Database) Close() error {
	return d.driver.Close()
}
