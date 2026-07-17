//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func newMemoryDataStore(t *testing.T) *DataStore {
	db, openErr := sql.Open("sqlite3", ":memory:")
	if openErr != nil {
		t.Fatalf("failed to open an in-memory database: %v", openErr)
		return nil
	}
	t.Cleanup(func() { _ = db.Close() })
	return &DataStore{db: db}
}

// busyErr mimics how a SQLITE_BUSY failure ("database is locked" in the
// mattn/go-sqlite3 driver) surfaces from the persist layer: wrapped in
// an SDKError chain.
func busyErr() *sdkErrors.SDKError {
	return sdkErrors.ErrEntityQueryFailed.Wrap(
		errors.New("database is locked"),
	)
}

func TestTransientDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "raw busy error",
			err:  errors.New("database is locked"),
			want: true,
		},
		{
			name: "raw locked error",
			err:  errors.New("database table is locked"),
			want: true,
		},
		{
			name: "SDKError-wrapped busy error",
			err:  busyErr(),
			want: true,
		},
		{
			name: "other sqlite error",
			err:  errors.New("UNIQUE constraint failed: policies.name"),
			want: false,
		},
		{
			name: "plain error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transientDBError(tt.err); got != tt.want {
				t.Errorf("transientDBError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithSerializableTx_RetriesTransientErrors(t *testing.T) {
	s := newMemoryDataStore(t)

	attempts := 0
	err := s.withSerializableTx(context.Background(), "test",
		func(_ *sql.Tx) *sdkErrors.SDKError {
			attempts++
			if attempts < 3 {
				return busyErr()
			}
			return nil
		})

	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
		return
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithSerializableTx_PermanentErrorDoesNotRetry(t *testing.T) {
	s := newMemoryDataStore(t)

	permanent := sdkErrors.ErrEntityInvalid.Clone()
	attempts := 0
	err := s.withSerializableTx(context.Background(), "test",
		func(_ *sql.Tx) *sdkErrors.SDKError {
			attempts++
			return permanent
		})

	if err == nil {
		t.Fatal("expected the permanent error to surface")
		return
	}
	if !err.Is(sdkErrors.ErrEntityInvalid) {
		t.Errorf("expected ErrEntityInvalid, got: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected exactly 1 attempt, got %d", attempts)
	}
}

func TestWithSerializableTx_TimeoutBoundsRetries(t *testing.T) {
	t.Setenv(env.NexusDBOperationTimeout, "300ms")

	s := newMemoryDataStore(t)

	attempts := 0
	err := s.withSerializableTx(context.Background(), "test",
		func(_ *sql.Tx) *sdkErrors.SDKError {
			attempts++
			return busyErr()
		})

	if err == nil {
		t.Fatal("expected an error once the operation deadline expired")
		return
	}
	if attempts < 2 {
		t.Errorf("expected at least 2 attempts before the deadline,"+
			" got %d", attempts)
	}
}
