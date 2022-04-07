#!/bin/bash
set -o errexit
set -o posix

source ./scripts/deploy-helpers.sh

ensure_kubectl
ensure_kubeconfig

echo ">> Undeploying the Artillery LoadTest operator"
kubectl delete -k config/default
