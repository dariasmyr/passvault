package storage

import "errors"

var (
	ErrEntryNotFound         = errors.New("entry not found")
	ErrEncryptionKeyNotFound = errors.New("encryption key not found")
)
