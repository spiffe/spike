//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package lifecycle provides utilities for managing bootstrap state in
// Kubernetes environments. It handles coordination between multiple bootstrap
// instances to ensure bootstrap operations run exactly once per cluster.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	k8s "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const k8sTrue = "true"
const k8sServiceAccountNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
const hostNameEnvVar = "HOSTNAME"

const keyBootstrapCompleted = "bootstrap-completed"
const keyBootstrapCompletedAt = "completed-at"
const keyBootstrapCompletedByPod = "completed-by-pod"

// ShouldBootstrap determines whether the bootstrap process should be
// skipped based on the current environment and state. The function follows
// this decision logic:
//
//  0. Skip for Lite backend and In-Memory backend
//  1. Else, if SPIKE_BOOTSTRAP_FORCE="true", always proceed (return true)
//  2. In bare-metal environments (non-Kubernetes), always proceed
//  3. In Kubernetes environments, check the "spike-bootstrap-state" ConfigMap:
//     - If ConfigMap exists and bootstrap-completed="true", skip bootstrap
//     - Otherwise, proceed with bootstrap
//
// The function returns false if bootstrap should be skipped, true if it
// should proceed.
func ShouldBootstrap() bool {
	const fName = "ShouldBootstrap"

	// Memory backend doesn't need bootstrap.
	if env.BackendStoreTypeVal() == env.Memory {
		log.Info(
			fName,
			"message", "skipping bootstrap for in-memory backend",
		)
		return false
	}

	// Lite backend doesn't need bootstrap.
	if env.BackendStoreTypeVal() == env.Lite {
		log.Info(
			fName,
			"message", "skipping bootstrap for lite backend",
		)
		return false
	}

	// Check if we're forcing the bootstrap
	if os.Getenv(env.BootstrapForce) == k8sTrue {
		log.Info(fName, "message", "force bootstrap enabled")
		return true
	}

	// Try to detect if we're running in Kubernetes
	// InClusterConfig looks for:
	// - KUBERNETES_SERVICE_HOST env var
	// - /var/run/secrets/kubernetes.io/serviceaccount/token
	cfg, cfgErr := rest.InClusterConfig()
	if cfgErr != nil {
		// We're not in Kubernetes (bare-metal scenario)
		// Bootstrap should proceed in non-k8s environments
		if errors.Is(cfgErr, rest.ErrNotInCluster) {
			log.Info(
				fName,
				"message",
				"not running in Kubernetes: proceeding with bootstrap",
			)
			return true
		}

		// Some other error. Skip bootstrap.
		failErr := sdkErrors.ErrK8sClientFailed.Clone()
		failErr.Msg = "failed to get Kubernetes config: skipping bootstrap"
		log.WarnErr(fName, *failErr)
		return false
	}

	// We're in Kubernetes---check the ConfigMap
	clientset, clientErr := kubernetes.NewForConfig(cfg)
	if clientErr != nil {
		failErr := sdkErrors.ErrK8sClientFailed.Clone()
		failErr.Msg = "failed to create Kubernetes client: skipping bootstrap"
		log.WarnErr(fName, *failErr)
		// Can't check state, skip bootstrap.
		return false
	}

	namespace := "spike"
	// Read namespace from the service account if not specified
	if nsBytes, readErr := os.ReadFile(k8sServiceAccountNamespace); readErr == nil {
		namespace = string(nsBytes)
	}

	cm, getErr := clientset.CoreV1().ConfigMaps(namespace).Get(
		context.Background(),
		env.BootstrapConfigMapNameVal(),
		k8sMeta.GetOptions{},
	)
	if getErr != nil {
		failErr := sdkErrors.ErrK8sReconciliationFailed.Wrap(getErr)
		// ConfigMap doesn't exist or can't read it - proceed with bootstrap
		failErr.Msg = "failed to get ConfigMap: proceeding with bootstrap"
		log.WarnErr(fName, *failErr)
		return true
	}

	bootstrapCompleted := cm.Data[keyBootstrapCompleted] == k8sTrue
	completedAt := cm.Data[keyBootstrapCompletedAt]
	completedByPod := cm.Data[keyBootstrapCompletedByPod]

	if bootstrapCompleted {
		reason := fmt.Sprintf(
			"completed at %s by pod %s",
			completedAt, completedByPod,
		)
		log.Info(
			fName,
			"message", "skipping bootstrap based on ConfigMap state",
			keyBootstrapCompletedAt, completedAt,
			keyBootstrapCompletedByPod, completedByPod,
			"reason", reason,
		)
		return false
	}

	// Bootstrap not completed: proceed with bootstrap
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
func MarkBootstrapComplete() *sdkErrors.SDKError {
	const fName = "MarkBootstrapComplete"

	// Only mark complete in Kubernetes environments
	config, cfgErr := rest.InClusterConfig()
	if cfgErr != nil {
		if errors.Is(cfgErr, rest.ErrNotInCluster) {
			// Not in Kubernetes, nothing to mark
			log.Info(
				fName,
				"message", "not in Kubernetes: skipping completion marker",
			)
			return nil
		}

		failErr := sdkErrors.ErrK8sReconciliationFailed.Clone()
		failErr.Msg = "failed to get Kubernetes config"
		return failErr.Wrap(cfgErr)
	}

	clientset, clientErr := kubernetes.NewForConfig(config)
	if clientErr != nil {
		failErr := sdkErrors.ErrK8sReconciliationFailed.Clone()
		failErr.Msg = "failed to create Kubernetes client"
		return failErr.Wrap(clientErr)
	}

	namespace := "spike"
	if nsBytes, readErr := os.ReadFile(k8sServiceAccountNamespace); readErr == nil {
		namespace = string(nsBytes)
	} else {
		failErr := sdkErrors.ErrK8sReconciliationFailed.Wrap(readErr)
		failErr.Msg = "failed to read service account namespace: using default: " +
			namespace
		log.WarnErr(fName, *failErr)
	}

	// Create ConfigMap marking bootstrap as complete
	cm := &k8s.ConfigMap{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: env.BootstrapConfigMapNameVal(),
		},
		Data: map[string]string{
			keyBootstrapCompleted:      k8sTrue,
			keyBootstrapCompletedAt:    time.Now().UTC().Format(time.RFC3339),
			keyBootstrapCompletedByPod: os.Getenv(hostNameEnvVar),
		},
	}

	ctx := context.Background()
	_, createErr := clientset.CoreV1().ConfigMaps(
		namespace,
	).Create(ctx, cm, k8sMeta.CreateOptions{})
	if createErr != nil {
		// Try to update if it already exists
		_, updateErr := clientset.CoreV1().ConfigMaps(
			namespace,
		).Update(ctx, cm, k8sMeta.UpdateOptions{})
		if updateErr != nil {
			failErr := sdkErrors.ErrK8sReconciliationFailed.Wrap(updateErr)
			failErr.Msg = "failed to mark bootstrap complete in ConfigMap"
			return failErr
		}
	}

	log.Info(
		fName,
		"message", "marked bootstrap as complete in ConfigMap",
	)
	return nil
}

// TODO: after replacing errors; check this file for logical consistency and documentation.
