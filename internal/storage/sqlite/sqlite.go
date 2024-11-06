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

// CreateEntry inserts a new entry into the vault table
func (s *Storage) CreateEntry(ctx context.Context, accountID int64, entryType, entryData string) (int64, error) {
	const op = "storage.sqlite.CreateEntry"
	query := `INSERT INTO vault (account_id, entry_type, entry_data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
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

// GetVault retrieves a vault from the vault table by vault ID
func (s *Storage) GetVault(ctx context.Context, vaultID int64) (*models.Vault, error) {
	const op = "storage.sqlite.GetVault"
	query := `SELECT id, account_id, entry_type, entry_data, created_at, updated_at FROM vault WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var vault models.Vault
	err = stmt.QueryRowContext(ctx, vaultID).Scan(&vault.ID, &vault.AccountId, &vault.EntryType, &vault.EntryData, &vault.CreatedAt, &vault.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrVaultNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &vault, nil
}

// UpdateVault updates an existing vault in the vault table by ID
func (s *Storage) UpdateVault(ctx context.Context, vaultID int64, vaultType, vaultData string) error {
	const op = "storage.sqlite.UpdateVault"
	query := `UPDATE vault SET entry_type = ?, entry_data = ?, updated_at = ? WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vaultType, vaultData, time.Now(), vaultID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// DeleteVault removes a vault from the vault table by ID
func (s *Storage) DeleteVault(ctx context.Context, vaultID int64) error {
	const op = "storage.sqlite.DeleteVault"
	query := `DELETE FROM vault WHERE id = ?`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vaultID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// ListVaults retrieves all vaults for a given user from the vault table
func (s *Storage) ListEntries(ctx context.Context, accountID int64) ([]*models.Vault, error) {
	const op = "storage.sqlite.ListEntries"
	query := `SELECT id, user_id, entry_type, entry_data, created_at, updated_at FROM vault WHERE account_id = ?`
	rows, err := s.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var vaults []*models.Vault
	for rows.Next() {
		var vault models.Vault
		if err := rows.Scan(&vault.ID, &vault.AccountId, &vault.EntryType, &vault.EntryData, &vault.CreatedAt, &vault.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		vaults = append(vaults, &vault)
	}
	return vaults, nil
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
