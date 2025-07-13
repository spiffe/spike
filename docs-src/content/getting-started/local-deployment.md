+++
# //    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Building SPIKE Locally"
weight = 3
sort_by = "weight"
+++

# Building SPIKE Locally and Deploying to Minikube

If you want to contribute to the **SPIKE** codebase and test your changes on a 
local Kubernetes cluster, follow this guide. If you want to build SPIKE from
the source code but want to test it on a *bare metal* Linux machine without 
using any containerization solution, check out 
[**SPIKE** on Linux][spike-on-linux] instead.

In this guide we will follow a similar approach to 
[**SPIKE** Quickstart][quickstart], with the following changes:

1. Build container images locally from existing source code.
2. Push the container images to a local container registry. 
3. Use a customized `values.yaml` for the helm charts to create a more
   production-like namespace structure.

[spike-on-linux]: @/getting-started/bare-metal.md "SPIKE on Linux"
[quickstart]: @/getting-started/quickstart.md "Getting Started with SPIKE"

Without further ado, let's begin with the prerequisites.

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

## Forwarding the Local Registry

Since Minikube contains its Kubernetes cluster in a virtual machine, accessing
its container registry requires network mapping. There are several ways to 
achieve this. In this guide, we will use Kubernetes port forwarding.

Open a terminal, change your directory to the project's folder and execute
the following:

```bash 
make docker-forward-registry
# Sample Output:
# Forwarding from 127.0.0.1:5000 -> 5000
# Forwarding from [::1]:5000 -> 5000
```

Leave this terminal open. We will execute the rest of the commands on 
a separate terminal window.

## Build Container Images Locally

We have a `make` target to build the container images locally.

```bash
make docker-build
```

## Pushing Container Images to the Local Registry

Next up, we'll push the container images to our internal Minikube container
registry:

```bash
# Make sure you port forward using 
# `make docker-forward-registry`
# before executing this command:
make docker-push
```

## Deploying SPIRE and SPIKE to the Local Cluster

Once we push the container images to the registry, we can now deploy **SPIRE**
and **SPIKE**.

```bash
# Uses `./config/helm/values-custom.yaml`
make deploy-local
```

## Verifying SPIKE Deployment

First, make sure that your components are up and running.

The following commands should **all** show `Ready` and `Runing` containers.

```bash
kubectl get po -n spire-server
# spiffe-server-0
# spiffe-spiffe-oidc-discovery-provider
kubectl get po -n spire-system
# spiffe-agent
# spiffe-spiffe-csi-driver
kubectl get po -n spike
# spiffe-spike-keeper-0
# spiffe-spike-keeper-1
# spiffe-spike-keeper-2
# spiffe-spike-nexus-0
# spiffe-spike-pilot
```

You can also shell into **SPIKE Pilot** to create and retrieve secrets to 
ensure **SPIKE** is up and running and properly configured in the cluster.

```bash
make exec-spike
# ^ This will shell into SPIKE Pilot.
# Now you can execute the following commands:
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

## You Are All Set

That's it. Now, you know how to modify **SPIKE**'s source code and test your
changes in a local Kubernetes cluster.

Next up, you might want to [Read **SPIKE**'s Source Code][github] to learn more
about **SPIKE**'s internals, or [learn more about **SPIKE**'s 
architecture][architecture] or [security model][security-model].

You might also want to try [building **SPIKE** on a bare metal 
Linux][bare-metal], if you want to see how **SPIKE** can be used on a bare
metal Linux machine without using container orchestration such as *Kubernetes*


[github]: https://github.com/spiffe/spike "SPIKE on GitHub"
[architecture]: @/architecture/_index.md "SPIKE Architecture"
[security-model]: @/architecture/security-model.md "SPIKE Security Model"
[bare-metal]: @/getting-started/bare-metal.md "SPIKE on Linux"

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
