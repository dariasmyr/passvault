package models

import "time"

type Entry struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	AccountId int64
	EntryType string
	EntryData string
}
