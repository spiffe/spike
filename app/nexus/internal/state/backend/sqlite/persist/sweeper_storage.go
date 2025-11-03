//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"fmt"

	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/kek"
)

// ListSecretsWithKEK returns all secrets using a specific KEK
func (s *DataStore) ListSecretsWithKEK(ctx context.Context, kekID string) ([]kek.SecretPath, error) {
	const fName = "ListSecretsWithKEK"

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx,
		`SELECT path, version FROM secrets WHERE kek_id = ? ORDER BY path, version`,
		kekID)
	if err != nil {
		log.Log().Error(fName,
			"message", "failed to query secrets",
			"kek_id", kekID,
			"err", err.Error())
		return nil, fmt.Errorf("%s: %w", fName, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Log().Error(fName, "message", "failed to close rows", "err", closeErr.Error())
		}
	}()

	var secrets []kek.SecretPath
	for rows.Next() {
		var sp kek.SecretPath
		if err := rows.Scan(&sp.Path, &sp.Version); err != nil {
			log.Log().Error(fName, "message", "failed to scan row", "err", err.Error())
			continue
		}
		secrets = append(secrets, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iteration error: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "listed secrets with KEK",
		"kek_id", kekID,
		"count", len(secrets))

	return secrets, nil
}
