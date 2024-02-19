.PHONY: all

all: log

log:
	go run . log --not-redmine --range day
