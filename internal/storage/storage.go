package storage

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrVaultNotFound         = errors.New("vault not found")
	ErrEncryptionKeyNotFound = errors.New("encryption key not found")
)
