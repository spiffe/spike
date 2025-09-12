//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lifecycle provides utilities for managing bootstrap state in Kubernetes
// environments. It handles coordination between multiple bootstrap instances
// to ensure bootstrap operations run exactly once per cluster.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
	k8s "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	appEnv "github.com/spiffe/spike/app/bootstrap/internal/env"
)

const k8sTrue = "true"
const k8sServiceAccountNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// ShouldBootstrap determines whether the bootstrap process should be
// skipped based on the current environment and state. The function follows
// this decision logic:
//
//  1. If SPIKE_BOOTSTRAP_FORCE="true", always proceed (return true)
//  2. In bare-metal environments (non-Kubernetes), always proceed
//  3. In Kubernetes environments, check the "spike-bootstrap-state" ConfigMap:
//     - If ConfigMap exists and bootstrap-completed="true", skip bootstrap
//     - Otherwise, proceed with bootstrap
//
// The function returns false if bootstrap should be skipped, true if it
// should proceed.
func ShouldBootstrap() bool {
	const fName = "bootstrap.shouldSkipBootstrap"

	// Memory backend doesn't need bootstrap.
	if appEnv.BackendStoreType() == appEnv.Memory {
		log.Log().Info(fName,
			"message", "Skipping bootstrap for 'in memory' backend")
		return false
	}

	// Lite backend doesn't need bootstrap.
	if appEnv.BackendStoreType() == appEnv.Lite {
		log.Log().Info(fName,
			"message", "Skipping bootstrap for 'lite' backend")
		return false
	}

	// Check if we're forcing bootstrap
	if os.Getenv(env.BootstrapForce) == k8sTrue {
		log.Log().Info(fName, "message", "Force bootstrap enabled")
		return true
	}

	// Try to detect if we're running in Kubernetes
	// InClusterConfig looks for:
	// - KUBERNETES_SERVICE_HOST env var
	// - /var/run/secrets/kubernetes.io/serviceaccount/token
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// We're not in Kubernetes (bare-metal scenario)
		// Bootstrap should proceed in non-k8s environments
		if errors.Is(err, rest.ErrNotInCluster) {
			log.Log().Info(fName,
				"message", "Not running in Kubernetes, proceeding with bootstrap",
			)
			return true
		}

		// Some other error. Skip bootstrap.
		log.Log().Error(fName,
			"message",
			"Could not determine cluster config. Skipping bootstrap",
			"err", err.Error())
		return false
	}

	// We're in Kubernetes - check the ConfigMap
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Log().Error(fName,
			"message",
			"Failed to create Kubernetes client. SKIPPING bootstrap.",
			"err", err.Error())
		// Can't check state, skip bootstrap.
		return false
	}

	namespace := "spike"
	// Read namespace from the service account if not specified
	if nsBytes, err := os.ReadFile(k8sServiceAccountNamespace); err == nil {
		namespace = string(nsBytes)
	}

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(
		context.Background(),
		appEnv.ConfigMapName(),
		k8sMeta.GetOptions{},
	)
	if err != nil {
		// ConfigMap doesn't exist or can't read it - proceed with bootstrap
		log.Log().Info(fName,
			"message",
			"ConfigMap not found or not readable, proceeding with bootstrap",
			"err", err.Error())
		return true
	}

	bootstrapCompleted := cm.Data["bootstrap-completed"] == k8sTrue
	completedAt := cm.Data["completed-at"]
	completedByPod := cm.Data["completed-by-pod"]

	if bootstrapCompleted {
		reason := fmt.Sprintf("completed at %s by pod %s",
			completedAt, completedByPod)
		log.Log().Info(fName,
			"message", "Skipping bootstrap based on ConfigMap state",
			"completed-at", completedAt,
			"completed-by-pod", completedByPod,
			"reason", reason,
		)
		return false
	}

	// Boostrap not completed---proceed with bootstrap
	return true
}

// MarkBootstrapComplete creates or updates the "spike-bootstrap-state"
// ConfigMap in Kubernetes to mark the bootstrap process as successfully
// completed. The ConfigMap includes:
//
//   - bootstrap-completed: "true"
//   - completed-at: RFC3339 timestamp
//   - completed-by-pod: hostname of the pod that completed bootstrap
//
// This function only operates in Kubernetes environments. In bare-metal
// deployments, it logs a message and returns nil without error.
//
// If the ConfigMap already exists, it will be updated. If creation fails,
// an update operation is attempted as a fallback.
func MarkBootstrapComplete() error {
	const fName = "bootstrap.markBootstrapComplete"

	// Only mark complete in Kubernetes environments
	config, err := rest.InClusterConfig()
	if err != nil {
		if errors.Is(err, rest.ErrNotInCluster) {
			// Not in Kubernetes, nothing to mark
			log.Log().Info(fName,
				"message", "Not in Kubernetes, skipping completion marker")
			return nil
		}
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	namespace := "spike"
	if nsBytes, err := os.ReadFile(k8sServiceAccountNamespace); err == nil {
		namespace = string(nsBytes)
	}

	// Create ConfigMap marking bootstrap as complete
	cm := &k8s.ConfigMap{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: appEnv.ConfigMapName(),
		},
		Data: map[string]string{
			"bootstrap-completed": k8sTrue,
			"completed-at":        time.Now().UTC().Format(time.RFC3339),
			"completed-by-pod":    os.Getenv("HOSTNAME"),
		},
	}

	ctx := context.Background()
	_, err = clientset.CoreV1().ConfigMaps(
		namespace,
	).Create(ctx, cm, k8sMeta.CreateOptions{})
	if err != nil {
		// Try to update if it already exists
		_, err = clientset.CoreV1().ConfigMaps(
			namespace,
		).Update(ctx, cm, k8sMeta.UpdateOptions{})
	}

	if err != nil {
		log.Log().Error(fName,
			"message", "Failed to mark bootstrap complete", "err", err.Error())
		return err
	}

	log.Log().Info(fName,
		"message", "Marked bootstrap as complete in ConfigMap")
	return nil
}
