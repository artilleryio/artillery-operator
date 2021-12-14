#!/bin/sh
set -o errexit

kubectl_cmd='kubectl'

is_kubeconfig_missing () {
   if [ ! -f "$1" ]; then
       echo "kubeconfig [$1] is missing!"
       echo "Please setup your kubeconfig to allow access to your Kubernetes cluster."
       exit 1
   fi
}

if ! kubectl_loc="$(type -p "$kubectl_cmd")" || [[ -z $kubectl_loc ]]; then
  echo "kubectl is missing!"
  echo "Please install kubectl and try again, https://kubernetes.io/docs/tasks/tools/#kubectl"
  exit 1
fi

if [[ -z "${KUBECONFIG}" ]]; then
  is_kubeconfig_missing "${HOME}/.kube/config"
else
  echo "KUBECONFIG is setup with ${KUBECONFIG}"
  echo
  is_kubeconfig_missing "${KUBECONFIG}"
fi

echo ">> Undeploying the Artillery LoadTest operator"
kubectl delete -k config/default
