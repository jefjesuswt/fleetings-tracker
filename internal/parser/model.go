package parser

import "time"

type Reminder struct {
	ID      string
	File    string
	Content string
	DueDate time.Time
}
