+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Quickstart"
weight = 3
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

[docker]: https://www.docker.com/ "Docker: Build, Share, and Run Applications"
[kubectl]: https://kubernetes.io/docs/tasks/tools/ "kubectl: Kubernetes command-line tool"
[make]: https://www.gnu.org/software/make/ "GNU Make: Build Automation Tool"

## Starting Minikube

To start a local Minikube cluster, clone the project repository and run the
following command in the root directory of the project:

```bash
cd $WORKSPACE # Replace with your workspace directory
git clone https://github.com/spiffe/spike.git
cd spike
make minikube-start # This will start a Minikube cluster.
```

If successful, you will have a local Minikube cluster running with the
necessary plugins enabled. You can verify that Minikube is running by executing:

```bash
minikube status
# or...
kubectl get node
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
helm repo add spiffe https://spiffe.github.io/helm-charts-hardened
helm repo update

helm upgrade --install -n spire-mgmt spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  --create-namespace
  
helm upgrade --install -n spire-mgmt spiffe \
  https://spiffe.github.io/helm-charts-hardened \
  -f ./values.yaml # The values.yaml file we created earlier
```

## Verifying SPIKE Deployment

Once the deployment is complete, you can verify SPIKE is running by 
creating a sample secret and reading its value back.

```bash
kubectl exec -it -n spire-mgmt deploy/spire-spike-pilot -- sh
# Shell into the container and run the following commands:
spike secret list
spike secret put test/creds username=spike password=SPIKERocks
spike secret list
spike secret get test/creds
```

## Next Up

You are all set. You have successfully deployed SPIKE to your local Minikube
cluster. Explore other parts of the documentation to learn more about
using SPIKE.

Here are a few links to get you started:

* Building SPIKE Locally and Deploying to Minikube
* Bare Metal SPIKE Installation
* Configuring SPIKE
* SPIKE Architecture
* SPIKE Production Hardening Guide
