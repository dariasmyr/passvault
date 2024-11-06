package models

import "time"

type EncryptionKey struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	AccountId int64
	KeyPart   string
}
