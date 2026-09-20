//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"context"
	stdnet "net"
	"testing"
	"time"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func TestWaitForKeepers_ReturnsOnceEveryKeeperListens(t *testing.T) {
	listener, listenErr := stdnet.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatalf("failed to open a listener: %v", listenErr)
		return
	}
	t.Cleanup(func() {
		if closeErr := listener.Close(); closeErr != nil {
			t.Errorf("failed to close the listener: %v", closeErr)
		}
	})

	keepers := map[string]string{
		"1": "https://" + listener.Addr().String(),
	}
	if waitErr := waitForKeepers(
		context.Background(), keepers, 5*time.Second,
	); waitErr != nil {
		t.Fatalf("expected no error for a listening keeper, got %v", waitErr)
	}
}

func TestWaitForKeepers_FailsWhenAKeeperNeverListens(t *testing.T) {
	// Reserve a port and close it so nothing listens there.
	listener, listenErr := stdnet.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatalf("failed to reserve a port: %v", listenErr)
		return
	}
	address := listener.Addr().String()
	if closeErr := listener.Close(); closeErr != nil {
		t.Fatalf("failed to release the port: %v", closeErr)
	}

	keepers := map[string]string{"1": "https://" + address}
	waitErr := waitForKeepers(context.Background(), keepers, 3*time.Second)
	if waitErr == nil {
		t.Fatal("expected an error for a keeper that never listens")
		return
	}
	if waitErr.Code != sdkErrors.ErrBootstrapKeeperUnreachable.Code {
		t.Errorf("unexpected error code %q", waitErr.Code)
	}
}

func TestKeeperAddress(t *testing.T) {
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{"https://keeper-0.headless:8443", "keeper-0.headless:8443", false},
		{"https://keeper.example.org", "keeper.example.org:443", false},
		{"not a url", "", true},
		{"", "", true},
	}
	for _, c := range cases {
		got, gotErr := keeperAddress(c.in)
		if (gotErr != nil) != c.wantErr {
			t.Errorf("%q: error = %v, wantErr %v", c.in, gotErr, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}
