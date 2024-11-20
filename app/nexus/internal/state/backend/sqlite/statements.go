//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

const queryInitialize = `
CREATE TABLE IF NOT EXISTS admin_token (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	nonce BLOB NOT NULL,
	encrypted_token BLOB NOT NULL,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_recovery_metadata (
    id INTEGER PRIMARY KEY CHECK (id = 1),
	encrypted_root_key BLOB NOT NULL,
    token_hash BLOB NOT NULL,
    salt BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
`

const queryUpdateSecretMetadata = `
INSERT INTO secret_metadata (path, current_version, created_time, updated_time)
VALUES (?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
	current_version = excluded.current_version,
	updated_time = excluded.updated_time
`

const queryUpsertSecret = `
INSERT INTO secrets (path, version, nonce, encrypted_data, created_time, deleted_time)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(path, version) DO UPDATE SET
	nonce = excluded.nonce,
	encrypted_data = excluded.encrypted_data,
	deleted_time = excluded.deleted_time
`

const querySecretMetadata = `
SELECT current_version, created_time, updated_time 
FROM secret_metadata 
WHERE path = ?
`

const querySecretVersions = `
SELECT version, nonce, encrypted_data, created_time, deleted_time 
FROM secrets 
WHERE path = ?
ORDER BY version
`

const queryInsertAdminToken = `
INSERT INTO admin_token (id, nonce, encrypted_token, updated_at)
VALUES (1, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
	nonce = excluded.nonce,
	encrypted_token = excluded.encrypted_token,
	updated_at = excluded.updated_at
`

const querySelectAdminSigningToken = `
SELECT nonce, encrypted_token 
FROM admin_token 
WHERE id = 1
`
