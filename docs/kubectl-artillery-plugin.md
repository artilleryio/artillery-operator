# Kubectl Artillery plugin

The plugin mainly contains helper commands to speed creating native Kubernetes LoadTests.

- [Installation](#installation)
- [Sub commands](#sub-commands)

## Installation

### Download the binary

Download the binary from [GitHub Releases](https://github.com/artilleryio/artillery-operator/releases) and drop it in
your `$PATH`.

#### Linux

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/kubectl-artillery/releases/download/v0.1.0/kubectl-artillery_v0.1.0_linux_amd64.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(amd64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/kubectl-artillery/releases/download/v0.1.0/kubectl-artillery_v0.1.0_darwin_amd64.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(arm64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/kubectl-artillery/releases/download/v0.1.0/kubectl-artillery_v0.1.0_darwin_arm64.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

### Verify installation

You can verify its installation using `kubectl`:

```shell
$ kubectl plugin list
#The following kubectl-compatible plugins are available:

# /usr/local/bin/kubectl-artillery
```

Validate if `kubectl artillery` can be executed.

```bash
kubectl artillery --help
# Use artillery.io operator helpers

# Usage:
#  artillery [flags]
#  artillery [command]

# Available Commands:
#  completion  generate the autocompletion script for the specified shell
#  generate    Generates load test manifests and wires dependencies in a kustomization.yaml file
#  help        Help about any command

# Flags:
#  -h, --help   help for artillery

# Use "artillery [command] --help" for more information about a command.
```

## Sub commands

- [generate](#generate-load-tests)

### generate

Use the `generate` sub command to generate a load test manifest, related Kubernetes manifests (e.g. ConfigMap). All
wired in a kustomization.yaml file.

- [Example: generate and apply Load Test](#example-generate-and-apply-load-test)

```shell
kubectl artillery generate --help
# Generates load test manifests and wires dependencies in a kustomization.yaml file

# Usage:
#  artillery generate [OPTIONS] [flags]

# Aliases:
#  generate, gen

# Examples:
...

# Flags:
#  -s, --script string   Specify path to artillery test-script file
#  -e, --env string      Optional. Specify the load test environment - defaults to dev (default "dev")
#  -o, --out string      Optional. Specify output path to write load test manifests and kustomization.yaml
#  -c, --count int       Optional. Specify number of load test workers (default 1)
#  -h, --help            help for generate
```

#### Output is configurable

By default, all manifests will be written to an `artillery-manifests` directory. This will be located in the same path
where the `generate` sub command was run.

Use the `--out/-o` flag to specify a different directory path to write load test manifests and kustomization.yaml.

#### Test scripts are bundled

The `generate` sub command also copies the artillery test-script file to the output directory. This is
because [Kustomize v2.0 added a security check](https://kubectl.docs.kubernetes.io/faq/kustomize/#security-file-foo-is-not-in-or-below-bar)
that prevents kustomizations from reading files outside their own directory root.

### Example: generate and apply Load Test

```shell
kubectl artillery gen boom -s hack/examples/basic-loadtest/test-script.yaml
# artillery-manifests/loadtest-cr.yaml generated
# artillery-manifests/kustomization.yaml generated
```

Looking into the `artillery-manifests` directory reveals the generated manifests and bundled a copy of
the `test-script.yaml`
file.

```shell
ll ./artillery-manifests
# total 24
# drwx------   5 xxx  xxx   160B 18 Mar 15:40 .
# drwxr-xr-x@ 34 xxx  xxx   1.1K 22 Mar 17:28 ..
# -rw-r--r--   1 xxx  xxx   316B 23 Mar 14:11 kustomization.yaml
# -rw-r--r--   1 xxx  xxx   177B 23 Mar 14:11 loadtest-cr.yaml
# -rw-r--r--   1 xxx  xxx   805B 23 Mar 14:11 test-script.yaml
```

You can edit the files as you please. And finally apply the LoadTest to an Artillery Operator enabled cluster.

```shell
kubectl apply -k ./artillery-manifests
# configmap/boom-test-script created
# loadtest.loadtest.artillery.io/boom created
```

The `generate` sub command has configured the Kustomization.yaml file with a `configMapGenerator`. When applied, it has
generated `configmap/boom-test-script` which contains your Artillery test-script.

```shell
kubectl describe configmap/boom-test-script
# Name:         boom-test-script
# Namespace:    default
# Labels:       artillery.io/component=loadtest-config
#               artillery.io/part-of=loadtest
# Annotations:  <none>

# Data
# ====
# test-script.yaml:
# ----
# # In Artillery, each VU will be assigned to one of the defined
# ...
# config:
#   target: "http://prod-publi-bf4z9b3ymbgs-1669151092.eu-west-1.elb.amazonaws.com:8080"
#      ...

# scenarios:
#   - name: "Access the / route"
#     ...
# BinaryData
# ====
# 
# Events:  <none>
```
