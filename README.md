# artillery-operator

## Running locally

### Pre-requisites

- [Go installed](https://golang.org/doc/install).
- [Docker Desktop](https://docs.docker.com/desktop/#download-and-install) up and running.
- [KinD installed](https://kind.sigs.k8s.io/docs/user/quick-start#installation), `brew install kind` on macOS.

### Overview

The instructions here will help you set up, develop and deploy the operator locally. You will need the following:

- A local Kubernetes cluster on [KinD](https://kind.sigs.k8s.io) to host and run the operator.
- A [local docker registry](https://docs.docker.com/registry/) to store the operator image for deployment.
- To get comfortable with the `make` commands required to update and deploy the operator.

### Create a KinD cluster and local docker registry

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

### Add local registry domain to /etc/hosts

Append below to your `/etc/hosts` file

```text
# Added to resolve the local docker KinD registry domain
127.0.0.1 kind-registry
# End of section
```

### Local development and deployment

#### Modifying the *_types.go

After modifying the *_types.go file always run the following command to update the generated code for that resource
type:

```shell
make generate
```

#### Updating CRD and other manifests

CRD, RBAC and other manifests can be generated and updated with the following command:

```shell
make manifests
```

These manifests are located in the `config` directory.

#### Local development

You can run the operator as a Go program outside of the cluster. This method is useful for development purposes to speed
up deployment and testing.

The following command installs the CRDs in the cluster configured in your `~/.kube/config` file and runs the Operator as
a Go program locally:

```shell
make install run
```

#### Local deployment

A new namespace is created with name <project-name>-system, ex. artillery-operator-system, and will be used for the
deployment.

The following command will build and push an operator image to the local registry `kind-registry` tagged as
`kind-registry:5000/artillery-operator:v0.0.1`:

```shell
make docker-build docker-push IMG=kind-registry:5000/artillery-operator:<version>
```

Then, run the following to deploy the operator to the K8s cluster specified in `~/.kube/config`. This will also install
the RBAC manifests from config/rbac.

```shell
make deploy IMG=kind-registry:5000/artillery-operator:v0.0.1
```

## Sample LoadTest CR

This can be applied to a K8s artillery operator enabled cluster.

Save yaml below to file. Then apply:

```shell
kubectl apply -f path/to/loadtest.yaml
```

```yaml
apiVersion: loadtest.artillery.io/v1alpha1
kind: LoadTest
metadata:
  name: loadtest-sample
  namespace: load-tester
spec:
  # Add fields here
  count: 10
  environment: stage
  testScript:
    config:
      configMap: load-test-config
    external:
      payload:
        configMaps:
          - csv-payload-1
          - csv-payload-2
          - csv-payload-3
      processor:
        main:
          configMap: my-functions-js
        related:
          configMaps:
            - package-json
            - helper-js
```
