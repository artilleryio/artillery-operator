#!/bin/bash
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
set -o posix

# full directory name of the script no matter where it is being called from
# shellcheck disable=SC2039
script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# create kube-prometheus directory to download the operator - ignored by git and docker
kube_prom_dir="${script_dir}/kube-prometheus-main"

# uninstall the pushgateway
helm uninstall prometheus-pushgateway
helm repo remove prometheus-community

# delete the monitoring stack
kubectl delete --ignore-not-found=true -f "${kube_prom_dir}/manifests/" -f "${kube_prom_dir}/manifests/setup"
