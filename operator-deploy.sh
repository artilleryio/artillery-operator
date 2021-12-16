#!/bin/sh
set -o errexit

source ./scripts/deploy-helpers.sh

ensure_kubectl
ensure_kubeconfig

echo ">> Deploying the Artillery LoadTest operator"
kubectl apply -k config/default
