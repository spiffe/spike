//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"context"
	"fmt"
	stdnet "net"
	"net/url"
	"time"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// keeperProbeInterval is the pause between connection attempts while
// waiting for a Keeper to start listening.
const keeperProbeInterval = 2 * time.Second

// waitForKeepers blocks until every Keeper accepts a TCP connection or the
// per-Keeper timeout expires. It sends nothing; a Keeper that accepts a
// connection has its SVID and is ready to take a contribution.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - keepers: Keeper IDs mapped to their API root URLs.
//   - timeout: The longest wait per Keeper; zero or negative waits forever.
//
// Returns:
//   - *sdkErrors.SDKError: ErrBootstrapKeeperUnreachable naming the first
//     Keeper that did not start listening in time, or ErrDataInvalidInput
//     for a Keeper URL without a usable host; nil when all listen.
func waitForKeepers(
	ctx context.Context, keepers map[string]string, timeout time.Duration,
) *sdkErrors.SDKError {
	const fName = "waitForKeepers"

	for keeperID, keeperURL := range keepers {
		address, addrErr := keeperAddress(keeperURL)
		if addrErr != nil {
			return addrErr
		}

		keeperCtx := ctx
		cancel := context.CancelFunc(func() {})
		if timeout > 0 {
			keeperCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		log.Info(fName, "message", "waiting for keeper to listen",
			"keeper_id", keeperID, "address", address)
		listening := waitForAddress(keeperCtx, address)
		cancel()
		if !listening {
			failErr := sdkErrors.ErrBootstrapKeeperUnreachable.Clone()
			failErr.Msg = fmt.Sprintf(
				"keeper %s at %s did not start listening within %s; "+
					"no shares were sent, so bootstrap is safe to rerun",
				keeperID, address, timeout,
			)
			return failErr
		}
	}
	return nil
}

// waitForAddress dials an address until a connection is accepted or the
// context ends.
//
// Parameters:
//   - ctx: Context bounding the wait.
//   - address: The host:port to dial.
//
// Returns:
//   - bool: True when a connection was accepted before the context ended.
func waitForAddress(ctx context.Context, address string) bool {
	dialer := stdnet.Dialer{Timeout: keeperProbeInterval}
	for {
		conn, dialErr := dialer.DialContext(ctx, "tcp", address)
		if dialErr == nil {
			if closeErr := conn.Close(); closeErr != nil {
				log.Warn("waitForAddress", "message",
					"failed to close probe connection", "err", closeErr.Error())
			}
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(keeperProbeInterval):
		}
	}
}

// keeperAddress extracts the host:port to probe from a Keeper API root URL,
// defaulting the port to 443 when the URL carries none.
//
// Parameters:
//   - keeperURL: The Keeper API root, for example https://keeper:8443.
//
// Returns:
//   - string: The host:port to dial.
//   - *sdkErrors.SDKError: ErrDataInvalidInput when the URL has no host.
func keeperAddress(keeperURL string) (string, *sdkErrors.SDKError) {
	parsed, parseErr := url.Parse(keeperURL)
	if parseErr != nil || parsed.Hostname() == "" {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "keeper URL has no usable host: " + keeperURL
		return "", failErr
	}
	port := parsed.Port()
	if port == "" {
		port = "443"
	}
	return stdnet.JoinHostPort(parsed.Hostname(), port), nil
}
