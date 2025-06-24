![SPIKE](../../assets/spike-banner-lg.png)

## Prerequisites

The setup assumes four machines with the following network configuration.

You can change your test cluster to your liking, but, in that case, you 
will need to modify some scripts, as there might be hard-coded values
in the configuration files:

* `./config/helm/values-demo-edge-1.yaml`
* `./config/helm/values-demo-edge-2.yaml`
* `./config/helm/values-demo-mgmt.yaml`
* `./config/helm/values-demo-workload.yaml`

### Management

This machine will contain a **Minikube** Kubernetes cluster that has

* **SPIKE Nexus**
* and **SPIKE Pilot**

There will also be **SPIRE** deployed on the cluster as the Identity Control
Plane, federated with the **Workload**, **Edge 1**, **Edge 2** clusters.

Machine details:

```text
      inet: 10.211.55.28  
   netmask: 255.255.255.0  
 broadcast: 10.211.55.255
       DNS: spiffe-mgmt-cluster.shared
Linux User: mgmt
  Hostname: mgmt
   Host OS: Ubuntu 24.04.2 LTS
```

### Edge 1

This machine will contain a **Minikube** Kubernetes cluster that has
three **SPIKE Keeper** instances

There will also be **SPIRE** deployed on the cluster as the Identity Control
Plane, federated with the **Management** cluster.

Machine details:

```text
     inet: 10.211.55.26
  netmask: 255.255.255.0
broadcast: 10.211.55.255
      DNS: spiffe-edge-cluster-1.shared
Linux User: mgmt
  Hostname: mgmt
   Host OS: Ubuntu 24.04.2 LTS
```

### Edge 2

This machine will contain a **Minikube** Kubernetes cluster that has
three **SPIKE Keeper** instances.

There will also be **SPIRE** deployed on the cluster as the Identity Control
Plane, federated with the **Management** cluster.

Machine details:

```text
     inet: 10.211.55.27 
  netmask: 255.255.255.0  
broadcast: 10.211.55.255
      DNS: spiffe-edge-cluster-2.shared
Linux User: edge-2
  Hostname: edge-2
   Host OS: Ubuntu 24.04.2 LTS
```

### Workload

This machine will contain a **Minikube** Kubernetes cluster that has
a demo workload that can securely receive secrets from the Management
cluster.

There will also be **SPIRE** deployed on the cluster as the Identity Control
Plane, federated with the **Management** cluster.

Machine details:

```text
     inet: 10.211.55.25
  netmask: 255.255.255.0
broadcast: 10.211.55.255
      DNS: spiffe-workload-cluster.shared
Linux User: workload
  Hostname: workload
   Host OS: Ubuntu 24.04.2 LTS
```

### Apps Required

You will the following binaries installed on all machines:

* `git`
* `make`
* `docker`
* `kubectl`
* `minikube`
* `tmux` (optional, but helpful)

## Steps

> **$WORKSPACE**
> 
> Hint: `$WORKSPACE` in the following examples is the folder that
> you have cloned <https://github.com/spiffe/spike>.

### Clone SPIKE Source Code

Run this in all machines:

```bash
cd $WORKSPACE
git clone https://github.com/spiffe/spike.git
```

### (Optional) Prune Docker

Run this in all machines:

```bash
cd $WORKSPACE/spike
make docker-cleanup
```

### Reset Minikube

Run this in all machines:

```bash
cd $WORKSPACE/spike
make k8s-reset
```

### Build Container Images

Run this in all machines:

```bash
cd $WORKSPACE/spike
make docker-build
```

### Forward Docker Registry

On every machine, run the following:

```bash
cd $WORKSPACE/spike
make docker-forward-registry
```

### Push Container Images to Local Registry

On every machine, run the following:

```bash
cd $WORKSPACE/spike
make docker-push
```

### Deploy the Demo Setup

This will deploy **SPIRE** and relevant **SPIKE** components to the machine.

The script selects what to deploy based on the machines **hostname**, so
makes sure your machine has the corresponding `hostname` defined in the
**Prerequisites** section above.

Run this in all machines:

```bash
cd $WORKSPACE/spike
make demo-deploy
```

### Port-Forward SPIRE Bundle Endpoints

In a production setup, there will likely be a kind of DNS, and Ingress 
Controller to establish network connectivity; however, since this is a 
demo setup, we will just `kubectl port-forward` to keep configuration
minimal.

Run the following at a separate terminal in all machines:

```bash
cd $WORKSPACE/spike
make demo-spire-bundle-port-forward
```

### Exchange Trust Bundles

To establish **federation**, **SPIRE Server**s need to establish initial
trust. There are several ways of doing this. The easiest way is to manually
exchange trust bundles.

To do that, execute the following in all machines providing credentials when
the script asks you:

```bash
cd $WORKSPACE/spike
make demo-bundle-extract
```

When that's done, execute the following the finish bundle exchange:

```bash
cd $WORKSPACE/spike
make demo-bundle-set
```

### Port Forward Nexus and Keepers

Again, in a production deployment, this will likely be done through an 
IngressController. We will use port forwarding for simplicity:

In the **Management** machine, Execute the following on a separate terminal:

```bash
cd $WORKSPACE/spike
make demo-spike-nexus-port-forward
```

In **Edge 1** and **Edge 2** clusters, execute the following on a separate
terminal:

```bash
cd $WORKSPACE/spike
make demo-spike-keeper-port-forward
```

### Deploy the Demo Workload

Let's deploy a demo app to check if we can receive secrets.

Execute this on the **Workload** machine.

```bash
cd $WORKSPACE/spike
make demo-deploy-workload
```

You can find the source code of this demo workload
at `./app/demo/cmd/main.go`.

### Set Policies to Enable the Workload to Fetch Secrets

For the workload to be able to create and fetch secrets, we need to
create policies to allow these actions.

Execute the following in the **Management** machine:

```bash
cd $WORKSPACE/spike
make demo-set-policies
```

### Fetch Secrets

Now, you should be able to fetch secrets on the sample workload.

Execute the following in the **Workload** machine:

```bash
cd $WORKSPACE/spike
make demo-workload-exec
# Sample Output:
# Secret found:
# password: SPIKE_Rocks
# username: SPIKE
```

And that finalizes the demo where we securely share secrets using **SPIFFE**
and **SPIKE**.
