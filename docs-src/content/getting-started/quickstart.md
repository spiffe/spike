+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Quickstart Guide"
weight = 2
sort_by = "weight"
+++

# SPIKE Quickstart Guide

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

## Deploying SPIKE to Minikube

Once you have Minikube running, you can deploy SPIKE to it from SPIFFE
helm charts.

First create a `values.yaml` file to enable SPIKE components:

```yaml
spike-keeper:
  enabled: true
spike-nexus:
  enabled: true
spike-pilot:
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

## Verifying SPIKE Deployment

First, make sure that your components are up and running.

```bash
kubectl get po 
# Sample Output:
#
# NAME                                      READY  STATUS 
# spiffe-agent-j6448                        1/1    Running
# spiffe-server-0                           2/2    Running
# spiffe-spiffe-csi-driver-ngqnk            2/2    Running
# spiffe-spiffe-oidc-discovery-provider-78  2/2    Running
# spiffe-spike-keeper-0                     1/1    Running
# spiffe-spike-keeper-1                     1/1    Running
# spiffe-spike-keeper-2                     1/1    Running
# spiffe-spike-nexus-0                      1/1    Running
# spiffe-spike-pilot-6997997fcb-nlqk8       1/1    Running
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

You are all set. You have successfully deployed SPIKE to your local Minikube
cluster. Explore other parts of the documentation to learn more about
using SPIKE.

Here are a few links to get you started:

* [Building SPIKE Locally and Deploying to Minikube](@/development/local-deployment.md)
* [Bare Metal SPIKE Installation](@/development/bare-metal.md)
* [Configuring SPIKE](@/usage/configuration.md)
* [SPIKE Architecture](@/architecture/_index.md)
* [SPIKE Production Hardening Guide](@/operations/production.md)

<p>&nbsp;</p>

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
