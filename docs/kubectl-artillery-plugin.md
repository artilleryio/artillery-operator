# Kubectl Artillery plugin

- [Installation](#installation)
- [Usage](#usage)

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

## Usage
