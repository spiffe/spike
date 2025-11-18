package persist_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Policy represents a security policy with both plaintext and encrypted fields
type Policy struct {
	ID                   string
	Name                 string
	SPIFFEIDPattern      string
	PathPattern          string
	Permissions          []string
	Nonce                []byte
	EncryptedSPIFFEID    []byte
	EncryptedPathPattern []byte
	EncryptedPermissions []byte
}

// LiteDataStore provides encrypted storage for policies using SQLite and AES-GCM
type LiteDataStore struct {
	db     *sql.DB
	Cipher cipher.AEAD
}

// NewLiteDataStore creates a new encrypted data store with SQLite backend
func NewLiteDataStore(path string, key []byte) (*LiteDataStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create policies table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS policies (
            id TEXT PRIMARY KEY,
            name TEXT,
            nonce BLOB,
            encrypted_spiffeid BLOB,
            encrypted_path BLOB,
            encrypted_permissions BLOB
        )
    `)
	if err != nil {
		return nil, err
	}

	return &LiteDataStore{
		db:     db,
		Cipher: aesgcm,
	}, nil
}

// GenerateNonce creates a random nonce for AES-GCM encryption
func (s *LiteDataStore) GenerateNonce() ([]byte, error) {
	nonce := make([]byte, s.Cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

// Encrypt encrypts data using AES-GCM with the provided nonce
func (s *LiteDataStore) Encrypt(nonce, data []byte) []byte {
	return s.Cipher.Seal(nil, nonce, data, nil)
}

// StorePolicy encrypts and stores a policy in the database
func (s *LiteDataStore) StorePolicy(ctx context.Context, p *Policy) error {
	// Generate a unique nonce for this policy
	nonce, err := s.GenerateNonce()
	if err != nil {
		return err
	}

	// Encrypt sensitive fields using the generated nonce
	p.Nonce = nonce
	p.EncryptedSPIFFEID = s.Encrypt(nonce, []byte(p.SPIFFEIDPattern))
	p.EncryptedPathPattern = s.Encrypt(nonce, []byte(p.PathPattern))
	p.EncryptedPermissions = s.Encrypt(nonce, []byte(fmt.Sprintf("%v", p.Permissions)))

	// Insert encrypted policy into database
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO policies(id, name, nonce, encrypted_spiffeid, encrypted_path, encrypted_permissions)
         VALUES(?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Nonce, p.EncryptedSPIFFEID, p.EncryptedPathPattern, p.EncryptedPermissions,
	)
	return err
}

// GetPolicy retrieves an encrypted policy from the database by ID
func (s *LiteDataStore) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, nonce, encrypted_spiffeid, encrypted_path, encrypted_permissions
         FROM policies WHERE id = ?`, id)

	p := &Policy{}
	err := row.Scan(&p.ID, &p.Name, &p.Nonce, &p.EncryptedSPIFFEID, &p.EncryptedPathPattern, &p.EncryptedPermissions)
	return p, err
}

// Decrypt decrypts ciphertext using AES-GCM with the provided nonce
func (s *LiteDataStore) Decrypt(nonce, ciphertext []byte) ([]byte, error) {
	return s.Cipher.Open(nil, nonce, ciphertext, nil)
}

// TestPolicyStoreSQLite tests the encrypted policy storage functionality
func TestPolicyStoreSQLite(t *testing.T) {
	ctx := context.Background()

	// Generate a random 32-byte key for AES-256
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	tmpfile := "test_policies.db"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpfile)

	// Create new encrypted data store
	ds, err := NewLiteDataStore(tmpfile, key)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test policy
	policy := &Policy{
		ID:              uuid.New().String(),
		Name:            "Test Policy",
		SPIFFEIDPattern: "spiffe://example.org/service",
		PathPattern:     "/service/*",
		Permissions:     []string{"read", "write"},
	}

	// Store the policy (this will encrypt sensitive fields)
	if err := ds.StorePolicy(ctx, policy); err != nil {
		t.Fatal(err)
	}

	// Retrieve the stored policy
	stored, err := ds.GetPolicy(ctx, policy.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Decrypt and verify the SPIFFE ID pattern
	decrypted, err := ds.Decrypt(stored.Nonce, stored.EncryptedSPIFFEID)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != policy.SPIFFEIDPattern {
		t.Fatalf("Decrypted value mismatch: got %s, want %s", decrypted, policy.SPIFFEIDPattern)
	}

	t.Logf("Policy stored and decrypted successfully: %s", decrypted)
}
