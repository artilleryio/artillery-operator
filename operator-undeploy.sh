#!/bin/bash
set -o errexit

source ./scripts/deploy-helpers.sh

ensure_kubectl
ensure_kubeconfig

echo ">> Undeploying the Artillery LoadTest operator"
kubectl delete -k config/default
