+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Quickstart Guide"
weight = 2
sort_by = "weight"
+++

# **SPIKE** Quickstart Guide

The fastest way to get started with **SPIRE** and **SPIKE** is to deploy them 
using the official [SPIFFE Helm chart][helm-charts-hardened].

[helm-charts-hardened]: https://github.com/spiffe/helm-charts-hardened "SPIFFE Helm charts (hardened)"

You can deploy SPIKE to any Kubernetes cluster, including a local one like
[KinD][kind] or [Minikube][minikube]. We will use **Minikube** in this guide.
Your installation may vary slightly depending on the Kubernetes cluster you
are using, but the general steps will be the same.

We will also use a [Debian Linux][debian] machine throughout this guide, but you
can use any OS that supports SPIFFE, SPIRE, Docker, and Kubernetes. Depending on
your OS, your installation steps may vary slightly, but the general steps will
not change much.

[kind]: https://kind.sigs.k8s.io/ "KinD: Kubernetes in Docker"
[minikube]: https://minikube.sigs.k8s.io/ "Minikube: Run Kubernetes locally"
[debian]: https://www.debian.org/ "Debian: The Universal Operating System"

## Prerequisites

Here is a list of things you need to have installed on your machine before
starting with this guide:

* Have [Docker][docker] installed and running on your machine.
* Have a [`kubectl`][kubectl] client installed. 
* Have [`make`][make] installed on your machine.
* Have a [`minikube`][minikube] binary installed.
* Have [`helm`][helm] binary installed.

[docker]: https://www.docker.com/ "Docker: Build, Share, and Run Applications"
[kubectl]: https://kubernetes.io/docs/tasks/tools/ "kubectl: Kubernetes command-line tool"
[make]: https://www.gnu.org/software/make/ "GNU Make: Build Automation Tool"
[helm]: https://helm.sh/ "Helm"

## Starting Minikube

To start a local Minikube cluster, clone the project repository and run the
following command in the root directory of the project:

```bash
cd $WORKSPACE # Replace with your workspace directory
git clone https://github.com/spiffe/spike.git
cd spike
make docker-cleanup # (Optional) Purge docker registry.
make k8s-delete     # Delete any former Minikube installation.
make k8s-start      # This will start a Minikube cluster.
```

If successful, you will have a local Minikube cluster running with the
necessary plugins enabled. You can verify that Minikube is running by executing:

```bash
minikube status
# or...
kubectl get node

# Sample Outputs:
#
# $ minikube status
# minikube
# type: Control Plane
# host: Running
# kubelet: Running
# apiserver: Running
# kubeconfig: Configured
#
# $ kubectl get node
# NAME       STATUS   ROLES           AGE   VERSION
# minikube   Ready    control-plane   67s   v1.33.1
```

## Deploying **SPIKE** to Minikube

Once you have Minikube running, you can deploy **SPIKE** to it from 
**SPIFFE helm charts**.

First create a `values.yaml` file to enable SPIKE components:

```yaml
# file: values.yaml

spike-keeper:
  enabled: true
  namespaceOverride: spike
  image:
    registry: ghcr.io
    repository: spiffe/spike-keeper
    pullPolicy: IfNotPresent
    tag: ""

spike-nexus:
  enabled: true
  namespaceOverride: spike
  image:
    registry: ghcr.io
    repository: spiffe/spike-nexus
    pullPolicy: IfNotPresent
    tag: ""

spike-pilot:
  enabled: true
  namespaceOverride: spike
  image:
    registry: ghcr.io
    repository: spike-pilot
    pullPolicy: IfNotPresent
    tag: ""
```

Then deploy SPIKE using the following command:

```bash 
helm upgrade --install spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/
  
helm upgrade --install spiffe spire \
  --repo https://spiffe.github.io/helm-charts-hardened \
  -f ./values.yaml # The values.yaml file we created earlier
```

## Bootstrapping **SPIKE Nexus**

To use **SPIKE Nexus**, we'll need to run a bootstrapper job that will
seed it with a secure random root key.

At the time of this writing, there is an ongoing work to automate this at
**SPIFFE Helm Charts** upstream repo; however, until that work is merged and
published, you'll need to create [the following `bootrsap.yaml` 
file][bootstrap-yaml] and apply it using `kubectl`.

The following YAML snippet has been slightly altered to fit into the 
documentation. This may cause parsing issues if you directly copy it from this 
page. If you want to use it, [check out `bootstrap.yaml` at 
GitHub][bootstrap-yaml] instead.

[bootstrap-yaml]: https://github.com/spiffe/spike/blob/main/hack/k8s/Bootstrap.yaml "Bootstrap.yaml"

```yaml
# file: bootstrap.yaml

apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app.kubernetes.io/instance: spiffe
    app.kubernetes.io/name: spike-bootstrap
  name: spiffe-spike-bootstrap
  namespace: spike
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: spiffe
        app.kubernetes.io/name: spike-bootstrap
        component: spike-bootstrap
    spec:
      restartPolicy: OnFailure
      containers:
        - name: spiffe-spike-bootstrap
          image: localhost:5000/spike-bootstrap:dev
          command: ["/bootstrap", "-init"]
          env:
            - name: SPIKE_NEXUS_API_URL
              value: https://spiffe-spike-nexus:443
            - name: SPIFFE_ENDPOINT_SOCKET
              value: unix:///spiffe-workload-api/spire-agent.sock
            - name: SPIKE_SYSTEM_LOG_LEVEL
              value: DEBUG
            - name: SPIKE_TRUST_ROOT
              value: spike.ist
            - name: SPIKE_NEXUS_SHAMIR_SHARES
              value: "3"
            - name: SPIKE_NEXUS_SHAMIR_THRESHOLD
              value: "2"
            - name: SPIKE_NEXUS_KEEPER_PEERS
              value: "spiffe-spike-keeper-0.spiffe-spike-keeper-headless:8443\
              ,https://spiffe-spike-keeper-1.spiffe-spike-keeper-headless:8443\
              ,https://spiffe-spike-keeper-2.spiffe-spike-keeper-headless:8443"
            - name: SPIKE_BOOTSTRAP_FORCE
              value: "false"
          imagePullPolicy: IfNotPresent
          resources: {}
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            seccompProfile:
              type: RuntimeDefault
          volumeMounts:
            - mountPath: /spiffe-workload-api
              name: spiffe-workload-api
              readOnly: true
      dnsPolicy: ClusterFirst
      securityContext:
        fsGroup: 1000
        fsGroupChangePolicy: OnRootMismatch
        runAsGroup: 1000
        runAsUser: 1000
      serviceAccountName: spiffe-spike-bootstrap
      volumes:
        - csi:
            driver: csi.spiffe.io
            readOnly: true
          name: spiffe-workload-api
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/instance: spiffe
    app.kubernetes.io/name: spike-bootstrap
  name: spiffe-spike-bootstrap
  namespace: spike
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: spike
  name: spiffe-bootstrap-role
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "update", "patch"]
    resourceNames: ["spike-bootstrap-state"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: spiffe-bootstrap-rolebinding
  namespace: spike
subjects:
  - kind: ServiceAccount
    name: spiffe-spike-bootstrap
    namespace: spike
roleRef:
  kind: Role
  name: spiffe-bootstrap-role
  apiGroup: rbac.authorization.k8s.io
```

With the above YAML file, execute the following:

```bash
kubectl apply -f bootstrap.yaml
```

## Verifying **SPIKE** Deployment

First, make sure that your components are up and running.

```bash
kubectl get po -A
# Sample Output:
#
# NAME                                              READY  STATUS 
# spike          spiffe-spike-bootstrap-x9nlr       1/1    Completed
# spike          spiffe-spike-keeper-0              1/1    Running
# spike          spiffe-spike-keeper-1              1/1    Running
# spike          spiffe-spike-keeper-2              1/1    Running
# spike          spiffe-spike-nexus-0               1/1    Running
# spike          spiffe-spike-pilot-5ddb88f-jsv9q   1/1    Running
# spire-server   spiffe-server-0                    2/2    Running
# spire-server   spiffe-oidc-provider-b4b9d-vn2zj   2/2    Running
# spire-system   spiffe-agent-lllsv                 1/1    Running
# spire-system   spiffe-spiffe-csi-driver-dkbwf     2/2    Running
```

Once the deployment is complete, you can verify SPIKE is running by 
creating a sample secret and reading its value back.

```bash
kubectl exec -it deploy/spiffe-spike-pilot -- sh
# Shell into the container and run the following commands:
spike secret list
# Output:
# No Secrets found.
spike secret put test/creds username=spike password=SPIKERocks
# Output:
# OK
spike secret list
# Output:
# - test/creds
spike secret get test/creds
# Output:
# password: SPIKERocks
# username: spike
```

## Next Up

You are all set. You have successfully deployed **SPIKE** to your local 
Minikube cluster. Explore other parts of the documentation to learn more about
using **SPIKE**.

Here are a few links to get you started:

* [Building **SPIKE** Locally and Deploying to Minikube][local-deployment]
* [Bare Metal **SPIKE** Installation][bare-metal]
* [Configuring **SPIKE**][configuration]
* [**SPIKE** Architecture][architecture]
* [*8SPIKE** Production Hardening Guide][hardening]
* [**SPIKE** CLI Reference][cli]

## Open Source Is Better Together

[**Join the SPIKE community**][community] to ask your questions and
learn from the subject-matter experts.

[hardening]: @/operations/production.md "SPIKE Production Hardening Guide"
[architecture]: @/architecture/_index.md "SPIKE Architecture"
[configuration]: @/usage/configuration.md "Configuring SPIKE"
[local-deployment]: @/development/local-deployment.md "Building SPIKE Locally and Deploying to Minikube"
[bare-metal]: @/development/bare-metal.md "Bare Metal SPIKE Installation"
[spiffe]: https://spiffe.io/ "Turtle Power"
[architecture]: @/architecture/system-overview.md "SPIKE Architecture"
[quickstart]: @/getting-started/quickstart.md "SPIKE Quickstart"
[community]: @/community/hello.md "Open Source is better together."
[cli]: @/usage/commands/_index.md
[github]: https://github.com/spiffe/spike

<p>&nbsp;</p>

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
