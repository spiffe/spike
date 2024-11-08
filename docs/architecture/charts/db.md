## SPIKE Database Usage

**SPIKE Nexus** is the only client for the backing store (*Postgres DB*).

Here are the things **SPIKE Nexus** stores in the db.

* **root key** (*encrypted with the admin password*)
* **admin token** (*encrypted with the root key*)
* **session keys** (*encrypted with the root key*)
* **secrets** (*encrypted with the root key*)

Note that both **admin token**, **session keys**, and **secrets**, are a kinds
of secrets from the data storage perspective.

Also note that **SPIKE Pilot** (i.e. `spike`) can save temporary session keys
and encrypted admin tokens on disk for convenience. Whereas **SPIKE Nexus**
will either store things in memory or keep them encrypted in a database, **never**
saving anything on the file system.

## SPIKE Data Model

Here is an initial data model (*subject to change during implementation*)

Note that for simplicity, we'll initially only support **Postgres** as a
backing store.

```sql
-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum for different types of secrets
CREATE TYPE secret_type AS ENUM ('admin_token', 'session_key', 'secret');

-- Store the root key (encrypted with admin password)
CREATE TABLE root_keys (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    encrypted_key bytea NOT NULL,
    key_hash bytea NOT NULL,  -- For verification
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    last_rotated_at timestamp with time zone,
    active boolean DEFAULT true,
    CONSTRAINT single_active_key UNIQUE (active)
);

-- Store all types of secrets (encrypted with root key)
CREATE TABLE secrets (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255) NOT NULL,
    type secret_type NOT NULL,
    encrypted_data bytea NOT NULL,
    metadata jsonb,  -- For additional secret-specific data
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone,
    version integer DEFAULT 1,
    previous_version_id uuid REFERENCES secrets(id),
    CONSTRAINT unique_active_name UNIQUE (name, type)
);

-- Audit log for all operations
CREATE TABLE audit_logs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation varchar(50) NOT NULL,
    secret_id uuid REFERENCES secrets(id),
    metadata jsonb,
    performed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_secrets_type ON secrets(type);
CREATE INDEX idx_secrets_name ON secrets(name);
CREATE INDEX idx_audit_logs_secret_id ON audit_logs(secret_id);
CREATE INDEX idx_audit_logs_performed_at ON audit_logs(performed_at);

-- Add triggers for updating 'updated_at' timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_secrets_updated_at
    BEFORE UPDATE ON secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```
