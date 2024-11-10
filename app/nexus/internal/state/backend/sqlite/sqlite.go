//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
)

func Register() {
	backend.Register("sqlite", New)
}

type Backend struct {
	db        *sql.DB
	cipher    cipher.AEAD
	mu        sync.RWMutex
	closeOnce sync.Once
	opts      *Options
}

func New(cfg backend.Config) (backend.Backend, error) {
	opts, err := parseOptions(cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("invalid sqlite options: %w", err)
	}

	key, err := hex.DecodeString(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	// Validate key length
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key length: must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Backend{
		cipher: gcm,
		opts:   opts,
	}, nil
}

func (s *Backend) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return errors.New("backend already initialized")
	}

	if err := s.createDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(s.opts.DataDir, s.opts.DatabaseFile)

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d",
		dbPath,
		s.opts.JournalMode,
		s.opts.BusyTimeoutMs))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(s.opts.MaxOpenConns)
	db.SetMaxIdleConns(s.opts.MaxIdleConns)
	db.SetConnMaxLifetime(s.opts.ConnMaxLifetime)

	// Create tables
	if err := s.createTables(ctx, db); err != nil {
		err := db.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to create tables: %w", err)
	}

	s.db = db
	return nil
}

func (s *Backend) createDataDir() error {
	return os.MkdirAll(s.opts.DataDir, 0750)
}

func (s *Backend) createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS admin_token (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			nonce BLOB NOT NULL,
			encrypted_token BLOB NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS secrets (
			path TEXT NOT NULL,
			version INTEGER NOT NULL,
			nonce BLOB NOT NULL,
			encrypted_data BLOB NOT NULL,
			created_time DATETIME NOT NULL,
			deleted_time DATETIME,
			PRIMARY KEY (path, version)
		);

		CREATE TABLE IF NOT EXISTS secret_metadata (
			path TEXT PRIMARY KEY,
			current_version INTEGER NOT NULL,
			created_time DATETIME NOT NULL,
			updated_time DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_secrets_path ON secrets(path);
		CREATE INDEX IF NOT EXISTS idx_secrets_created_time ON secrets(created_time);
	`)
	return err
}

func (s *Backend) encrypt(data []byte) ([]byte, []byte, error) {
	nonce := make([]byte, s.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := s.cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nonce, nil
}

func (s *Backend) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	plaintext, err := s.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return plaintext, nil
}

func (s *Backend) StoreSecret(ctx context.Context, path string, secret store.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false

	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	// Update metadata
	_, err = tx.ExecContext(ctx, `
		INSERT INTO secret_metadata (path, current_version, created_time, updated_time)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			current_version = excluded.current_version,
			updated_time = excluded.updated_time
	`, path, secret.Metadata.CurrentVersion, secret.Metadata.CreatedTime, secret.Metadata.UpdatedTime)
	if err != nil {
		return fmt.Errorf("failed to store secret metadata: %w", err)
	}

	// Update versions
	for version, sv := range secret.Versions {
		data, err := json.Marshal(sv.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal secret values: %w", err)
		}

		encrypted, nonce, err := s.encrypt(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret data: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO secrets (path, version, nonce, encrypted_data, created_time, deleted_time)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(path, version) DO UPDATE SET
				nonce = excluded.nonce,
				encrypted_data = excluded.encrypted_data,
				deleted_time = excluded.deleted_time
		`, path, version, nonce, encrypted, sv.CreatedTime, sv.DeletedTime)
		if err != nil {
			return fmt.Errorf("failed to store secret version: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true

	return nil
}

func (s *Backend) LoadSecret(ctx context.Context, path string) (*store.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var secret store.Secret

	// Load metadata
	err := s.db.QueryRowContext(ctx, `
		SELECT current_version, created_time, updated_time 
		FROM secret_metadata 
		WHERE path = ?
	`, path).Scan(&secret.Metadata.CurrentVersion, &secret.Metadata.CreatedTime, &secret.Metadata.UpdatedTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load secret metadata: %w", err)
	}

	// Load versions
	rows, err := s.db.QueryContext(ctx, `
		SELECT version, nonce, encrypted_data, created_time, deleted_time 
		FROM secrets 
		WHERE path = ?
		ORDER BY version
	`, path)
	if err != nil {
		return nil, fmt.Errorf("failed to query secret versions: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	secret.Versions = make(map[int]store.Version)
	for rows.Next() {
		var (
			version     int
			nonce       []byte
			encrypted   []byte
			createdTime time.Time
			deletedTime sql.NullTime
		)

		if err := rows.Scan(&version, &nonce, &encrypted, &createdTime, &deletedTime); err != nil {
			return nil, fmt.Errorf("failed to scan secret version: %w", err)
		}

		decrypted, err := s.decrypt(encrypted, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret version: %w", err)
		}

		var values map[string]string
		if err := json.Unmarshal(decrypted, &values); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret values: %w", err)
		}

		sv := store.Version{
			Data:        values,
			CreatedTime: createdTime,
		}
		if deletedTime.Valid {
			sv.DeletedTime = &deletedTime.Time
		}

		secret.Versions[version] = sv
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate secret versions: %w", err)
	}

	return &secret, nil
}

func (s *Backend) StoreAdminToken(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encrypted, nonce, err := s.encrypt([]byte(token))
	if err != nil {
		return fmt.Errorf("failed to encrypt admin token: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO admin_token (id, nonce, encrypted_token, updated_at)
		VALUES (1, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			nonce = excluded.nonce,
			encrypted_token = excluded.encrypted_token,
			updated_at = excluded.updated_at
	`, nonce, encrypted)
	if err != nil {
		return fmt.Errorf("failed to store admin token: %w", err)
	}

	return nil
}

func (s *Backend) LoadAdminToken(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nonce, encrypted []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT nonce, encrypted_token 
		FROM admin_token 
		WHERE id = 1
	`).Scan(&nonce, &encrypted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to load admin token: %w", err)
	}

	decrypted, err := s.decrypt(encrypted, nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt admin token: %w", err)
	}

	return string(decrypted), nil
}

func (s *Backend) Close(ctx context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	return err
}
