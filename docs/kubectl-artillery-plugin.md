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
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/releases/download/v0.1.1/kubectl-artillery_0.1.1_linux_amd64_2022-04-04T15.07.18Z.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(amd64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/releases/download/v0.1.1/kubectl-artillery_v0.1.1_darwin_amd64_2022-04-04T15.07.18Z.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(arm64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/artillery-operator/releases/download/v0.1.1/kubectl-artillery_v0.1.1_darwin_arm64_2022-04-04T15.07.18Z.tar.gz
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
#  ...
#  generate    Generates load test manifests configured in a kustomization.yaml file
#  ...
#  scaffold    Scaffolds test scripts from K8s services using liveness probe HTTP endpoints

# Flags:
#  -h, --help   help for artillery

# Use "artillery [command] --help" for more information about a command.
```

## Sub commands

- [scaffold](#scaffold)
- [generate](#generate)

### scaffold

Use the `scaffold` subcommand to
scaffold [test scripts](https://www.artillery.io/docs/guides/guides/test-script-reference)
from existing K8s [Services](https://kubernetes.io/docs/concepts/services-networking/service/).

Created test scripts use
the [expect plugin](https://www.artillery.io/docs/guides/plugins/plugin-expectations-assertions)
to functionally
test [HTTP liveness probe endpoints](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request)
in Pods proxied by the supplied services.

Once created, update the tests to match your requirements. You can also use created test scripts to generate load tests.

- [Example: scaffold test scripts](#example-scaffold-test-scripts)

#### Output is configurable

By default, all test scripts will be written to an `artillery-scripts` directory. This will be located in the same path
where the `scaffold` subcommand was run.

Use the `--out/-o` flag to specify a different directory path to write the test scripts.

#### A target url for every functional test

A Kubernetes Service may reference multiple ports, requiring multiple `target` urls. Created test scripts work around
this by using a full urls for every functional test endpoint.

For example,

```yaml
...
scenarios:
  - flow:
      - get:
          url: http://nginx-probes-mapped:80/
          expect:
            - statusCode: 200
...
```

#### Some liveness probes cannot be tested

A Pod may define a liveness probe on a port not accessible to the proxying Service. Such a liveness a probe cannot be
tested.

The plugin cannot scaffold a test script for a Service has no access to the proxied Pod's liveness probes.

### Example: scaffold test scripts

Let's check that our service `nginx-probes-mapped` has related Pods with defined HTTP liveness probes. A
service's `Selector` field helps us identify the correct Pods.

```shell
kubectl describe service nginx-probes-mapped
# Name:                     nginx-probes-mapped
# Namespace:                default
# ...
# Selector:                 app=nginx-probes-mapped <<<<<
# ...
# Port:                     nginx-http-port  80/TCP
# TargetPort:               80/TCP
# ...
```

```shell
kubectl get pods --selector=app=nginx-probes-mapped
# NAME                                 READY   STATUS    RESTARTS   AGE
# k8s-probes-mapped-64998cbdf5-7cqzg   1/1     Running   0          11m
```

```shell
kubectl get pods k8s-probes-mapped-64998cbdf5-7cqzg -o yaml
# apiVersion: v1
# kind: Pod
# ...
# spec:
#  containers:
#  - image: nginx
#    imagePullPolicy: Always
#    livenessProbe:
#      failureThreshold: 1
#      httpGet:
#        path: /
#        port: 80
#        scheme: HTTP
#      initialDelaySeconds: 1
...
```

The answer is YES. `nginx-probes-mapped` can access the Pod's HTTP liveness probe.

```shell
kubectl artillery scaffold nginx-probes-mapped
# artillery-scripts/test-script_nginx-probes-mapped.yaml generated
```

Looking into the `artillery-scripts` directory reveals the generated test script YAML file.

```shell
ls -alh ./artillery-scripts
# total 8
# drwx------   3 xxx  xxx    96B  1 Apr 15:22 .
# drwxr-xr-x@ 34 xxx  xxx   1.1K  1 Apr 15:22 ..
# -rw-r--r--   1 xxx  xxx   305B  4 Apr 13:20 test-script_nginx-probes-mapped.yaml
````

You can edit the files as you please. Then use it to generate a load test.

### generate

Use the `generate` subcommand to generate a load test manifest, related Kubernetes manifests (e.g. ConfigMap). All wired
in a kustomization.yaml file.

- [Example: generate and apply Load Test](#example-generate-and-apply-load-test)

#### Output is configurable

By default, all manifests will be written to an `artillery-manifests` directory. This will be located in the same path
where the `generate` subcommand was run.

Use the `--out/-o` flag to specify a different directory path to write load test manifests and kustomization.yaml.

#### Test scripts are bundled

The `generate` subcommand also copies the artillery test-script file to the output directory. This is
because [Kustomize v2.0 added a security check](https://kubectl.docs.kubernetes.io/faq/kustomize/#security-file-foo-is-not-in-or-below-bar)
that prevents kustomizations from reading files outside their own directory root.

### Example: generate and apply Load Test

```shell
kubectl artillery gen boom -s hack/examples/basic-loadtest/test-script.yaml
# artillery-manifests/loadtest-cr.yaml generated
# artillery-manifests/kustomization.yaml generated
```

Looking into the `artillery-manifests` directory reveals the generated manifests and bundled a copy of
the `test-script.yaml` file.

```shell
ls -alh ./artillery-manifests
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

The `generate` subcommand has configured the Kustomization.yaml file with a `configMapGenerator`. When applied, it has
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
