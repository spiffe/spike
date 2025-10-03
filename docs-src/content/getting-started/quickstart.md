+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Quickstart Guide"
weight = 2
sort_by = "weight"
+++

> **⚠️ Additional Instructions**
> 
> Since **SPIFFE Helm Charts** do not have **SPIKE Bootstrap** yet, the 
> instructions on this page have additional guidance to deploy **SPIKE** 
> using a local **SPIFFE Helm Charts** repo.
> 
> We will update this page once **SPIKE Bootstrap** is available in the 
> upstream **SPIFFE Helm Charts**.

[quickstart]: @/getting-started/quickstart.md "SPIKE Quickstart"

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

> ** ⚠️ Changes Due to Current Upstream Helm Charts Work**
> 
> There are some changes to the upstream **SPIFFE Helm Charts** that are
> currently in progress. Until they are merged, you will need to use a
> feature branch of the upstream repo.
> 
> For this, first clone the upstream repo:
> 
> ```bash
> git clone https://github.com/spiffe/helm-charts-hardened.git
> ```
> 
> Then, switch to the `spike-next` branch:
> 
> ```bash
> cd helm-charts-hardened
> git checkout spike-next
> ```
> 
> You can now use the `spike-next` branch of the upstream repo to deploy
> **SPIKE** to Minikube.
> 
> ```bash
> # $WORKSPACE is your local workspace directory where you cloned the 
> # helm-charts-hardened repo and the spike repo.
>
> # Create a new namespace for SPIRE components.
> kubectl create ns spire-mgmt
>
> # Deploy the CRDs.
> helm upgrade --install -n spire-mgmt "spire-crds" "spire-crds" \
> "WORKSPACE/helm-charts-hardened/charts/spire-crds" \
> --create-namespace
>
> # Deploy SPIRE and SPIKE components.
> helm upgrade --install -n spire-mgmt "spiffe" "spire" \
> "WORKSPACE/helm-charts-hardened/charts/spire" \
> -f /path/to/your/values.yaml
> ```

spife-helm-charts-hardened: https://spiffe.github.io/helm-charts-hardened/

Once you have Minikube running, you can deploy **SPIKE** to it from 
**SPIFFE helm charts**.

First create a `values.yaml` file to enable SPIKE components:

```yaml
# file: values.yaml
spike-nexus:
  enabled: true
spike-keeper:
  enabled: true
spike-pilot:
  enabled: true
spire-server:
  enabled: true
spire-agent:
  enabled: true
spiffe-csi-driver:
  enabled: true
spiffe-oidc-discovery-provider:
  enabled: true
```

Then deploy SPIKE using the following command:

```bash 
helm upgrade --install spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/
  
helm upgrade --install spiffe spire \
  --repo https://spiffe.github.io/helm-charts-hardened \
  -f ./values.yaml # The values.yaml file we created earlier
```

## Verifying **SPIKE** Deployment

First, make sure that your components are up and running.

```bash
kubectl get po -A
# Sample Output:
#
# NAME                                              READY  STATUS 
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
