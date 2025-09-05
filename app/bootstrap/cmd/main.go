//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/fips140"
	"flag"
	"fmt"
	"os"

	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/spiffe"
	svid "github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/internal/config"

	"github.com/spiffe/spike/app/bootstrap/internal/env"
	"github.com/spiffe/spike/app/bootstrap/internal/net"
	"github.com/spiffe/spike/app/bootstrap/internal/state"
	"github.com/spiffe/spike/app/bootstrap/internal/url"

	// Kubernetes client imports
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// shouldSkipBootstrap checks if bootstrap should be skipped based on:
// 1. Force bootstrap flag
// 2. Kubernetes ConfigMap state (if running in k8s)
// 3. Always proceeds in bare-metal environments
func shouldSkipBootstrap() (bool, string) {
	const fName = "bootstrap.shouldSkipBootstrap"

	// Check if we're forcing bootstrap
	if os.Getenv("SPIKE_FORCE_BOOTSTRAP") == "true" {
		log.Log().Info(fName, "message", "Force bootstrap enabled")
		return false, ""
	}

	// Try to detect if we're running in Kubernetes
	// InClusterConfig looks for:
	// - KUBERNETES_SERVICE_HOST env var
	// - /var/run/secrets/kubernetes.io/serviceaccount/token
	config, err := rest.InClusterConfig()
	if err != nil {
		// We're not in Kubernetes (bare-metal scenario)
		// Bootstrap should proceed in non-k8s environments
		if err == rest.ErrNotInCluster {
			log.Log().Info(fName, "message", "Not running in Kubernetes, proceeding with bootstrap")
			return false, ""
		}
		// Some other error - be conservative and proceed
		log.Log().Warn(fName,
			"message", "Could not determine cluster config, proceeding with bootstrap",
			"err", err.Error())
		return false, ""
	}

	// We're in Kubernetes - check the ConfigMap
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Log().Error(fName,
			"message", "Failed to create Kubernetes client, proceeding with bootstrap",
			"err", err.Error())
		// Can't check state, proceed with bootstrap to be safe
		return false, ""
	}

	// Get the namespace from environment or default to current namespace
	namespace := os.Getenv("SPIKE_NAMESPACE")
	if namespace == "" {
		// Read namespace from service account if not specified
		if nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			namespace = string(nsBytes)
		} else {
			namespace = "spike" // fallback to default
		}
	}

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(
		context.Background(),
		"spike-bootstrap-state",
		metav1.GetOptions{},
	)
	if err != nil {
		// ConfigMap doesn't exist or can't read it - proceed with bootstrap
		log.Log().Info(fName,
			"message", "ConfigMap not found or not readable, proceeding with bootstrap",
			"err", err.Error())
		return false, ""
	}

	skipBootstrap := cm.Data["skip-bootstrap"] == "true"
	reason := cm.Data["reason"]

	if skipBootstrap {
		log.Log().Info(fName,
			"message", "Skipping bootstrap based on ConfigMap state",
			"reason", reason,
			"last-check", cm.Data["last-check-time"])
	}

	return skipBootstrap, reason
}

func main() {
	const fName = "bootstrap.main"

	log.Log().Info(fName, "message", "Starting SPIKE bootstrap...", "version", config.BootstrapVersion)

	init := flag.Bool("init", false, "Initialize the bootstrap module")
	flag.Parse()
	if !*init {
		fmt.Println("")
		fmt.Println("Usage: bootstrap -init")
		fmt.Println("")
		os.Exit(1)
		return
	}

	// Check if we should skip bootstrap (Kubernetes state or bare-metal)
	skip, reason := shouldSkipBootstrap()
	if skip {
		log.Log().Info(fName,
			"message", "Bootstrap already completed previously. Skipping.",
			"reason", reason,
		)
		fmt.Println("Bootstrap already completed previously. Exiting.")
		os.Exit(0)
		return
	}

	src := net.Source()
	defer spiffe.CloseSource(src)
	sv, err := src.GetX509SVID()
	if err != nil {
		log.FatalLn(fName,
			"message", "Failed to get X.509 SVID",
			"err", err.Error())
		os.Exit(1)
		return
	}

	if !svid.IsBootstrap(env.TrustRoot(), sv.ID.String()) {
		log.Log().Error(
			"Authenticate: You need a 'bootstrap' SPIFFE ID to use this command.",
		)
		os.Exit(1)
		return
	}

	log.Log().Info(
		fName, "FIPS 140.3 enabled", fips140.Enabled(),
	)
	log.Log().Info(
		fName, "message", "Sending shards to SPIKE Keeper instances...",
	)
	for keeperID, keeperAPIRoot := range env.Keepers() {
		log.Log().Info(fName, "keeper ID", keeperID)
		net.Post(
			net.MTLSClient(src),
			url.KeeperEndpoint(keeperAPIRoot),
			net.Payload(
				state.KeeperShare(
					state.RootShares(), keeperID),
				keeperID,
			),
			keeperID,
		)
	}

	log.Log().Info(fName, "message", "Sent shards to SPIKE Keeper instances.")
	fmt.Println("Bootstrap completed successfully!")
	os.Exit(0)
}
