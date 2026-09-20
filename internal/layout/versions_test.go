//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package layout

import (
	"testing"
)

// TestSPIREVersionsHaveOneSource checks that every script which installs
// or builds SPIRE takes its versions from hack/lib/versions.sh and that
// the documentation pages showing those versions agree with it. A bump is
// then a single edit plus whatever doc lines this test names.
func TestSPIREVersionsHaveOneSource(t *testing.T) {
	root := moduleRoot(t)
	versions := spireVersions(t, root)

	for _, key := range []string{
		"SPIRE_HELM_CHART_VERSION",
		"SPIRE_CRDS_HELM_CHART_VERSION",
		"SPIRE_VERSION",
	} {
		if versions[key] == "" {
			t.Fatalf("hack/lib/versions.sh does not set %s", key)
		}
	}

	for _, script := range []string{
		"hack/k8s/spike-install.sh",
		"hack/k8s/spike-dev-install.sh",
		"hack/bare-metal/build/build-spire.sh",
		"ci/integration/minio-rolearn/setup.sh",
	} {
		fileContains(t, root, script, "lib/versions.sh")
	}

	fileContains(t, root, "docs-src/content/getting-started/quickstart.md",
		"--version "+versions["SPIRE_CRDS_HELM_CHART_VERSION"],
		"--version "+versions["SPIRE_HELM_CHART_VERSION"],
	)
	fileContains(t, root, "docs-src/content/development/bare-metal.md",
		"--branch "+versions["SPIRE_VERSION"],
	)
}
