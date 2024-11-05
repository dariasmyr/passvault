package models

import "time"

type User struct {
	ID           int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SSOAccountId int64
	Email        string
}
