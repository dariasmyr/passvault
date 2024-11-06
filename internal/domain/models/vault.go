package models

import "time"

type Vault struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	AccountId int64
	EntryType string
	EntryData string
}
