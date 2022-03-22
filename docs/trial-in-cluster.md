[![Generic badge](https://img.shields.io/badge/Stage-Early%20Alpha-red.svg)](https://shields.io/)

<img width="1012" alt="Kubernetes native load testing" src="../assets/artillery-operator-header.png">

# Deploy the Operator in your own cluster

This deploys the alpha operator image:
[artillery-operator-alpha](https://github.com/orgs/artilleryio/packages/container/package/artillery-operator-alpha).

## Pre-requisites

- [kubectl installed](https://kubernetes.io/docs/tasks/tools/#kubectl).
- [kubeconfig setup](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig) to access a
  cluster, either using the `KUBECONFIG` environment variable or `$HOME/.kube/config`.
- A local copy of the `artillery-operator` github repo,
  either [downloaded](https://github.com/artilleryio/artillery-operator/archive/refs/heads/main.zip) or
  using `git clone`.

## Deploy the operator

Ensure you can execute `operator-deploy.sh` found in the `artillery-operator` root directory.

Then simply run:

```shell
./operator-deploy.sh
```

This will install the operator image `ghcr.io/artilleryio/artillery-operator-alpha:latest` in your cluster. And, run it
from the `artillery-operator-system` namespace with restricted cluster permissions.

## Undeploy the operator

Ensure you can execute `operator-undeploy.sh` found in the `artillery-operator` root directory.

Then simply run:

```shell
./operator-undeploy.sh
```

This will remove the operator and any created namespaces, load tests, etc... from your cluster.
