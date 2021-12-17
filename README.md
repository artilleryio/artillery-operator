[![Generic badge](https://img.shields.io/badge/Stage-Early%20Alpha-red.svg)](https://shields.io/)

<img width="1012" alt="Kubernetes native load testing" src="assets/artillery-operator-header.png">

# Artillery Operator for Kubernetes

The Artillery Operator is an implementation of
a [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/#operators-in-kubernetes) that
enables Kubernetes native load testing.

## Trial in your own cluster

This deploys the alpha operator image:
[artillery-operator-alpha](https://github.com/orgs/artilleryio/packages/container/package/artillery-operator-alpha).

### Pre-requisites

- [kubectl installed](https://kubernetes.io/docs/tasks/tools/#kubectl).
- [kubeconfig setup](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig) to access a
  cluster, either using the `KUBECONFIG` environment variable or `$HOME/.kube/config`.
- A local copy of the `artillery-operator`,
  either [downloaded](https://github.com/artilleryio/artillery-operator/archive/refs/heads/main.zip) or
  using `git clone`.

### Deploy the operator

Ensure you can execute `operator-deploy.sh` found in the `artillery-operator` root directory.

Then simply run:

```shell
./operator-deploy.sh
```

This will install the operator image `ghcr.io/artilleryio/artillery-operator-alpha:latest` in your cluster. And, run it
from the `artillery-operator-system` namespace with restricted cluster permissions.

### Undeploy the operator

Ensure you can execute `operator-undeploy.sh` found in the `artillery-operator` root directory.

Then simply run:

```shell
./operator-undeploy.sh
```

This will remove the operator and any created namespaces, load tests, etc... from your cluster.

## Running Load Tests

### Pre-requisites

A cluster (remote or local) with artillery-operator [already deployed](#trial-in-your-own-cluster).

### Example

The example is available at `hack/basic-loadtest`.

It provides two load tests each configured with a different number of workers and a target Api to test.

The example includes a [`kustomize`](https://kustomize.io) manifest which generates the ConfigMap required to hold the
test script used by the load tests. The `kustomize` manifest will also apply the Load Test Custom Resource manifests to
your cluster.

```shell
kubectl apply -k hack/basic-loadtest

# configmap/test-script created
# loadtest.loadtest.artillery.io/basic-test created
# loadtest.loadtest.artillery.io/other-test created

kubectl get loadtests basic-test other-test
# NAME         COMPLETIONS   DURATION   AGE   ENVIRONMENT   IMAGE
# basic-test   0/2           55s        55s   dev           artilleryio/artillery:latest
# other-test   0/4           55s        55s   staging       artilleryio/artillery:latest
```

#### Test reports

We don't yet aggregate the test results for a Load Test. As such, you'll have to check the logs from each worker to
monitor its test reports.

You can use a LoadTests created published `Events` to do this. E.g. let's find `basic-test`'s workers.

```shell
kubectl describe loadtests basic-test

# ...
# ...
# Status:
# ...
# Events:
#  Type    Reason   Age   From                 Message
#  ----    ------   ----  ----                 -------
#  Normal  Created  25s   loadtest-controller  Created Load Test worker master job: basic-test
#  Normal  Running  25s   loadtest-controller  Running Load Test worker pod: basic-test-6w2rq
#  Normal  Running  25s   loadtest-controller  Running Load Test worker pod: basic-test-7fjxq
```

The `Events` section lists all the created workers. Using the first worker `basic-test-6w2rq`, we can follow its test
reports.

```shell
kubectl logs -f basic-test-6w2rq
```

Displays:

```shell

  Telemetry is on. Learn more: https://artillery.io/docs/resources/core/telemetry.html
Phase started: unnamed (index: 0, duration: 60s) 15:18:01(+0000)

--------------------------------------
Metrics for period to: 15:18:10(+0000) (width: 7.007s)
--------------------------------------

vusers.created_by_name.Access the / route: .................. 24
vusers.created.total: ....................................... 24
vusers.completed: ........................................... 24
...
...
--------------------------------------
Metrics for period to: 15:18:20(+0000) (width: 9.013s)
--------------------------------------
....
....
```

#### basic-test

The `basic-test` load test is created using the `hack/basic-loadtest/basic-test-cr.yaml` manifest.

```yaml
apiVersion: loadtest.artillery.io/v1alpha1
kind: LoadTest
metadata:
  name: basic-test
  namespace: default
  labels:
    "artillery.io/test-name": basic-test
    "artillery.io/component": loadtest
    "artillery.io/part-of": loadtest

spec:
  # Add fields here
  count: 2
  environment: dev
  testScript:
    config:
      configMap: test-script
```

It runs 2 workers against a test script loaded from `configmap/test-script`.

#### other-test

The `other-test` load test is created using the `hack/basic-loadtest/other-test-cr.yaml` manifest.

```yaml
apiVersion: loadtest.artillery.io/v1alpha1
kind: LoadTest
metadata:
  name: other-test
  namespace: default
  labels:
    "artillery.io/test-name": other-test
    "artillery.io/component": loadtest
    "artillery.io/part-of": loadtest

spec:
  # Add fields here
  count: 4
  environment: staging
  testScript:
    config:
      configMap: test-script
```

It runs 4 workers against a test script loaded from `configmap/test-script`.

## Developing

### With local deployment

#### Pre-requisites

- [Go installed](https://golang.org/doc/install).
- [Docker Desktop](https://docs.docker.com/desktop/#download-and-install) up and running.
- [KinD installed](https://kind.sigs.k8s.io/docs/user/quick-start#installation), `brew install kind` on macOS.

#### Overview

The instructions here will help you set up, develop and deploy the operator locally. You will need the following:

- A local Kubernetes cluster on [KinD](https://kind.sigs.k8s.io) to host and run the operator.
- A [local docker registry](https://docs.docker.com/registry/) to store the operator image for deployment.
- To get comfortable with the `make` commands required to update and deploy the operator.

#### Create a KinD cluster and local docker registry

We are going to create a cluster with one master, two worker nodes and one docker registry so that we can build, push
and deploy our operator into Kubernetes.

Ensure Docker Desktop is up and running. Then, execute the following:

```shell
# Run setup script 
chmod +x hack/kind/kind-with-registry.sh
./hack/kind/kind-with-registry.sh

# Ensure KinD is running
kind get nodes
kubectl get all --all-namespaces
```

#### Add local registry domain to /etc/hosts

Append below to your `/etc/hosts` file

```text
# Added to resolve the local docker KinD registry domain
127.0.0.1 kind-registry
# End of section
```

#### Local development and deployment

##### Modifying the *_types.go

After modifying the *_types.go file always run the following command to update the generated code for that resource
type:

```shell
make generate
```

##### Updating CRD and other manifests

CRD, RBAC and other manifests can be generated and updated with the following command:

```shell
make manifests
```

These manifests are located in the `config` directory.

##### Local development

You can run the operator as a Go program outside of the cluster. This method is useful for development purposes to speed
up deployment and testing.

The following command installs the CRDs in the cluster configured in your `~/.kube/config` file and runs the Operator as
a Go program locally:

```shell
make install run
```

##### Local deployment

A new namespace is created with name <project-name>-system, ex. artillery-operator-system, and will be used for the
deployment.

The following command will build and push an operator image to the local registry `kind-registry` tagged as
`kind-registry:5000/artillery-operator:v0.0.1`:

```shell
make docker-build docker-push IMG=kind-registry:5000/artillery-operator:v0.0.1
```

Then, run the following to deploy the operator to the K8s cluster specified in `~/.kube/config`. This will also install
the RBAC manifests from config/rbac.

```shell
make deploy IMG=kind-registry:5000/artillery-operator:v0.0.1
```

### With remote deployment

#### Pre-requisites

- [Go installed](https://golang.org/doc/install).
- [Docker Desktop](https://docs.docker.com/desktop/#download-and-install) up and running.
- [eksctl installed](https://eksctl.io/introduction/#installation) to set up a remote cluster
  on [AWS EKS](https://aws.amazon.com/eks/), `brew tap weaveworks/tap; brew install weaveworks/tap/eksctl` on macOS.

#### Overview

Use these instructions to deploy the operator remotely on an AWS EKS cluster. You will need the following:

- Either use or create (using `eksctl`) a remote Kubernetes cluster to host and run the operator. We're
  using [AWS EKS](https://aws.amazon.com/eks/) but feel free to use another provider.
- A remote container registry (e.g. Docker Hub, etc..) to store the operator image for deployment.
- To get comfortable with the `make` commands required to update and deploy the operator.

#### Create + configure access to a remote cluster

**Skip this if you already have a remote cluster ready.**

##### Create an EKS cluster

We'll be setting up a remote on AWS EKS using `eksctl`.

```shell
eksctl create cluster -f hack/aws/eksctl/cluster.yaml
```

This will create a cluster as specified in `hack/aws/eksctl/cluster.yaml` and should take around ~20 minutes.

##### Configure access to EKS cluster

We need to configure `kubeconfig` access to be able to deploy our operator into the cluster.

For that, we'll use the `aws` cli:

```shell
# if required, login using sso 
aws sso login

# create the kubeconfig file in a target location  
aws eks --region eu-west-1 update-kubeconfig --name es-cluster-1 --kubeconfig hack/aws/eksctl/kubeconfig

# configure access using the created or updated kubeconfig file 
export KUBECONFIG=hack/aws/eksctl/kubeconfig

# ensure the cluster is accessible
kubectl get nodes  # this should display 4 nodes running on aws e.g. ip-*.eu-west-1.compute.internal
```

Do NOT commit the `kubeconfig` file into source control as it's based on your own credentials.

#### Deploying to the remote cluster

The operator is deployed as a K8s `Deployment` resource that runs in a newly created namespace with cluster wide RBAC
permissions.

The Operator SDK simplifies this step by providing a set of tasks in the project's `Makefile`.

##### Step 1: Create and host container image remotely

**Note**: Ensure you have a remote registry (e.g. hub.docker.com, etc..) to host and share the operator's container
image with the cluster.

At Artillery, we
use [Github Packages (container registry)](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
as a remote registry. The hosted container will be publicly available meaning it does not require `imagePullSecrets`
config in the `Deployment` manifest.

```shell
export IMAGE_REPO_OWNER=ghcr.io/artilleryio
```

[Ensure you're authenticated](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)
, then build and push the image:

```shell
make docker-build docker-push
```

##### Step 2: Deploy to the remote cluster

Point to your `kubeconfig` file:

```shell
export KUBECONFIG=hack/aws/eksctl/kubeconfig
```

Deploy the operator:

```shell
make deploy

# namespace/artillery-operator-system created
# customresourcedefinition.apiextensions.k8s.io/loadtests.loadtest.artillery.io created
# serviceaccount/artillery-operator-controller-manager created
# role.rbac.authorization.k8s.io/artillery-operator-leader-election-role created
# clusterrole.rbac.authorization.k8s.io/artillery-operator-manager-role created
# clusterrole.rbac.authorization.k8s.io/artillery-operator-metrics-reader created
# clusterrole.rbac.authorization.k8s.io/artillery-operator-proxy-role created
# rolebinding.rbac.authorization.k8s.io/artillery-operator-leader-election-rolebinding created
# clusterrolebinding.rbac.authorization.k8s.io/artillery-operator-manager-rolebinding created
# clusterrolebinding.rbac.authorization.k8s.io/artillery-operator-proxy-rolebinding created
# configmap/artillery-operator-manager-config created
# service/artillery-operator-controller-manager-metrics-service created
# deployment.apps/artillery-operator-controller-manager created
```

This creates the necessary namespace,CRD, RBAC and artillery-operator-controller-manager to run the operator in the
remote cluster.

Ensure the operator is running correctly:

```shell
kubectl -n artillery-operator-system get pods # find the artillery-operator-controller-manager pod

kubectl -n artillery-operator-system logs artillery-operator-controller-manager-764f97bdc9-dgcm8 -c manager # view the logs to ensure all is well
```

## Known issues

### Duplicated test reports (rare)

Duplicate test reports happen when a restarted failed worker reports test run results the worker has previously reported
before it failed.

We mitigate for this inside the operator by creating Kubernetes Job resources configured with
`RestartPolicy=Never` when starting Pods and `BackoffLimit=0` for the actual Job. This attempts to run the Job once
without restarting failed Pods (workers) for almost all circumstances.

This should stop failed workers from restarting therefore stopping duplicate test reports.

**HOWEVER**, the communication between the job controller, apiserver, kubelet and containers is not atomic. So, it is
possible that under some very unlikely combination of crashes/restarts in the kubelet, containers and/or job controller,
that the program in the Job's pod template might be started twice.

Here's a basic example of how things can go wrong:

- A user manually deletes a Pod using `kubectl`.
- The Load Test's Job resource create another one Pod to restart the work, therefore duplicating test results. 

