// Main migrations init script: go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations
// Test migrations init script: go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./tests/migrations --migrations-table=migrations_test
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"passvault/internal/domain/models"
	"passvault/internal/storage"
	"time"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveEntry inserts a new entry into the Entry table
func (s *Storage) SaveEntry(ctx context.Context, accountID int64, entryType, entryData string) (int64, error) {
	const op = "storage.sqlite.CreateEntry"
	query := `INSERT INTO entry (account_id, entry_type, entry_data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, accountID, entryType, entryData, time.Now(), time.Now())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	entryID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return entryID, nil
}

// GetEntry retrieves a entry from the entry table by entry ID
func (s *Storage) GetEntry(ctx context.Context, entryID int64) (*models.Entry, error) {
	const op = "storage.sqlite.GetEntry"
	query := `SELECT id, account_id, entry_type, entry_data, created_at, updated_at FROM entry WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var entry models.Entry
	err = stmt.QueryRowContext(ctx, entryID).Scan(&entry.ID, &entry.AccountId, &entry.EntryType, &entry.EntryData, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrEntryNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entry, nil
}

// UpdateEntries updates an existing entry in the entry table by ID
func (s *Storage) UpdateEntry(ctx context.Context, entryID int64, entryType, entryData string) error {
	const op = "storage.sqlite.UpdateEntry"
	query := `UPDATE entry SET entry_type = ?, entry_data = ?, updated_at = ? WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, entryType, entryData, time.Now(), entryID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// DeleteEntries removes a entry from the entry table by ID
func (s *Storage) DeleteEntry(ctx context.Context, entryID int64) error {
	const op = "storage.sqlite.DeleteEntry"
	query := `DELETE FROM entry WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, entryID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// ListEntries retrieves all entries for a given userId from the entry table
func (s *Storage) ListEntries(ctx context.Context, accountID int64) ([]*models.Entry, error) {
	const op = "storage.sqlite.ListEntries"
	query := `SELECT id, user_id, entry_type, entry_data, created_at, updated_at FROM entry WHERE account_id = ?`
	rows, err := s.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		if err := rows.Scan(&entry.ID, &entry.AccountId, &entry.EntryType, &entry.EntryData, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}

// StoreKeyPart inserts a new key part for a user into the encryption_key table
func (s *Storage) StoreKeyPart(ctx context.Context, accountID int64, keyPart string) (int64, error) {
	const op = "storage.sqlite.StoreKeyPart"
	query := `INSERT INTO encryption_key (account_id, key_part, created_at, updated_at) VALUES (?, ?, ?, ?)`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, accountID, keyPart, time.Now(), time.Now())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	keyID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return keyID, nil
}

// RetrieveKeyPart retrieves a key part for an account from the encryption_key table
func (s *Storage) RetrieveKeyPart(ctx context.Context, accountID int64) (string, error) {
	const op = "storage.sqlite.RetrieveKeyPart"
	query := `SELECT key_part FROM encryption_key WHERE account_id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var keyPart string
	err = stmt.QueryRowContext(ctx, accountID).Scan(&keyPart)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrEncryptionKeyNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return keyPart, nil
}

// DeleteKeyPart removes a key part for an account from the encryption_key table
func (s *Storage) DeleteKeyPart(ctx context.Context, accountID int64) error {
	const op = "storage.sqlite.DeleteKeyPart"
	query := `DELETE FROM encryption_key WHERE account_id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, accountID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
