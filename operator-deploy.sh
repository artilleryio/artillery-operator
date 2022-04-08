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
#   http://mozilla.org/MPL/2.0/
#

set -o errexit
set -o posix

source ./scripts/deploy-helpers.sh

ensure_kubectl
ensure_kubeconfig

echo ">> Deploying the Artillery LoadTest operator"
kubectl apply -k config/default
echo ""
echo ">> Telemetry is enabled."
echo ">> We use that telemetry information to help us understand how Artillery is used."
echo ">> To help us prioritize work on new features and bug fixes, and to help us improve Artillery's performance and stability."
echo ""
echo ">> Run the command below if you choose to disable telemetry:"
echo ">> kubectl -n artillery-operator-system set env deployments/artillery-operator-controller-manager ARTILLERY_DISABLE_TELEMETRY=true"
