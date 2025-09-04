//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package ddl

// QueryInitialize defines the SQL schema for initializing the database.
// It includes table creation and index definitions for policies, secrets,
// and secret metadata. These tables handle secret storage, metadata, and
// policy management with relevant constraints and indices.
const QueryInitialize = `
CREATE TABLE IF NOT EXISTS policies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    nonce BLOB NOT NULL,  
    encrypted_spiffe_id_pattern BLOB NOT NULL,
    encrypted_path_pattern BLOB NOT NULL,
    encrypted_permissions BLOB NOT NULL,
    created_time INTEGER NOT NULL
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
	oldest_version INTEGER NOT NULL,
	created_time DATETIME NOT NULL,
	updated_time DATETIME NOT NULL,
	max_versions INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_secrets_path ON secrets(path);
CREATE INDEX IF NOT EXISTS idx_secrets_created_time ON secrets(created_time);
`

// QueryUpdateSecretMetadata is a SQL query for inserting or updating secret
// metadata. It updates the current version, oldest version, max versions, and
// updated time in conflict with the existing path.
const QueryUpdateSecretMetadata = `
INSERT INTO secret_metadata (path, current_version, oldest_version, 
  created_time, updated_time, max_versions)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
	current_version = excluded.current_version,
	oldest_version = excluded.oldest_version,
	updated_time = excluded.updated_time,
	max_versions = excluded.max_versions
`

// QueryUpsertSecret is a SQL query for inserting or updating the `secrets`
// records.
const QueryUpsertSecret = `
INSERT INTO secrets (path, version, nonce, encrypted_data, created_time, deleted_time)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(path, version) DO UPDATE SET
	nonce = excluded.nonce,
	encrypted_data = excluded.encrypted_data,
	deleted_time = excluded.deleted_time
`

// QuerySecretMetadata is a SQL query to fetch metadata of a secret by its path.
const QuerySecretMetadata = `
SELECT current_version, oldest_version, created_time, updated_time, max_versions
FROM secret_metadata 
WHERE path = ?
`

// QuerySecretVersions retrieves all versions of a secret from the database.
const QuerySecretVersions = `
SELECT version, nonce, encrypted_data, created_time, deleted_time 
FROM secrets 
WHERE path = ?
ORDER BY version
`

// QueryUpsertPolicy defines an SQL query to insert or update a policy record.
const QueryUpsertPolicy = `
INSERT INTO policies (
    id, 
    name, 
    nonce,
    encrypted_spiffe_id_pattern, 
    encrypted_path_pattern, 
    encrypted_permissions, 
    created_time
)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    nonce = excluded.nonce,
    encrypted_spiffe_id_pattern = excluded.encrypted_spiffe_id_pattern,
    encrypted_path_pattern = excluded.encrypted_path_pattern,
    encrypted_permissions = excluded.encrypted_permissions
`

// QueryDeletePolicy defines the SQL statement to delete a policy by its ID.
const QueryDeletePolicy = `
DELETE FROM policies 
WHERE id = ?
`

// QueryLoadPolicy is a SQL query to select policy details by ID
const QueryLoadPolicy = `
SELECT name, 
       encrypted_spiffe_id_pattern, 
       encrypted_path_pattern, 
       encrypted_permissions, 
       nonce, 
       created_time
FROM policies
WHERE id = ?
`

const QueryAllPolicies = `
SELECT id, 
       name, 
       encrypted_spiffe_id_pattern, 
       encrypted_path_pattern, 
       encrypted_permissions, 
       nonce, 
       created_time
FROM policies
`

const QueryPathsFromMetadata = `SELECT path FROM secret_metadata`
