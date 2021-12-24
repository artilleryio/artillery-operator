#!/bin/sh
#
# Copyright (c) 2022.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0.
#
# If a copy of the MPL was not distributed with
# this file, You can obtain one at
#
#     http://mozilla.org/MPL/2.0/
#

set -o errexit

# full directory name of the script no matter where it is being called from
# shellcheck disable=SC2039
script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# create kube-prometheus directory to download the operator - ignored by git and docker
kube_prom_dir="${script_dir}/kube-prometheus-main"
mkdir -p "${kube_prom_dir}"

# download and unpack the operator
curl -L \
 -v \
 'https://github.com/prometheus-operator/kube-prometheus/archive/refs/heads/main.zip' \
 -o "${script_dir}/kube-prometheus-main.zip"

tar -zxvf "${script_dir}/kube-prometheus-main.zip" -C "${script_dir}"
rm -f "${script_dir}/kube-prometheus-main.zip"

# source: https://prometheus-operator.dev/docs/prologue/quick-start/
# create the monitoring stack
kubectl apply --server-side -f "${kube_prom_dir}/manifests/setup"
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
kubectl apply -f "${kube_prom_dir}/manifests/"

# add the prometheus pushgateway repo to helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# source: https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-pushgateway
# install the pushgateway
helm install prometheus-pushgateway --atomic --set serviceMonitor.enabled=true prometheus-community/prometheus-pushgateway

# port-forward the prometheus-pushgateway service to localhost
kubectl port-forward svc/prometheus-pushgateway 9091
