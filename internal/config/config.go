//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const NexusVersion = "0.1.0"
const PilotVersion = "0.1.0"
const KeeperVersion = "0.1.0"

const NexusIssuer = "spike-nexus"
const NexusAdminSubject = "spike-admin"
const NexusAdminTokenId = "spike-admin-jwt"

func SpikePilotAdminTokenPath() string {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	// Create path for .spike folder
	spikeDir := filepath.Join(homeDir, ".spike")

	// Create directory if it doesn't exist
	err = os.MkdirAll(spikeDir, 0600)
	if err != nil {
		panic(err)
	}

	return filepath.Join(spikeDir, ".spike-admin.jwt")
}

// SpiffeEndpointSocket returns the UNIX domain socket address for the SPIFFE
// Workload API endpoint.
//
// The function first checks for the SPIFFE_ENDPOINT_SOCKET environment variable.
// If set, it returns that value. Otherwise, it returns a default development
//
//	socket path:
//
// "unix:///tmp/spire-agent/public/api.sock"
//
// For production deployments, especially in Kubernetes environments, it's
// recommended to set SPIFFE_ENDPOINT_SOCKET to a more restricted socket path,
// such as: "unix:///run/spire/agent/sockets/spire.sock"
//
// Default socket paths by environment:
//   - Development (Linux): unix:///tmp/spire-agent/public/api.sock
//   - Kubernetes: unix:///run/spire/agent/sockets/spire.sock
//
// Returns:
//   - string: The UNIX domain socket address for the SPIFFE Workload API endpoint
//
// Environment Variables:
//   - SPIFFE_ENDPOINT_SOCKET: Override the default socket path
func SpiffeEndpointSocket() string {
	p := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	if p != "" {
		return p
	}

	return "unix:///tmp/spire-agent/public/api.sock"
}
