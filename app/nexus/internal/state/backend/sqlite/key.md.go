//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

type DatabaseConfigKey string

const (
	KeyDataDir                DatabaseConfigKey = "data_dir"
	KeyDatabaseFile           DatabaseConfigKey = "database_file"
	KeyJournalMode            DatabaseConfigKey = "journal_mode"
	KeyBusyTimeoutMs          DatabaseConfigKey = "busy_timeout_ms"
	KeyMaxOpenConns           DatabaseConfigKey = "max_open_conns"
	KeyMaxIdleConns           DatabaseConfigKey = "max_idle_conns"
	KeyConnMaxLifetimeSeconds DatabaseConfigKey = "conn_max_lifetime_seconds"
)
