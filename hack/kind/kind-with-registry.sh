#!/bin/sh
set -o errexit

# full directory name of the script no matter where it is being called from
script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
echo "script dir is ${script_dir}"

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# create data directory to mount volumes - ignored by git and docker
mkdir -p ${script_dir}/data

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraMounts:
  - hostPath: ${script_dir}/data
    containerPath: /data
- role: worker
  extraMounts:
  - hostPath: ${script_dir}/data
    containerPath: /data
- role: worker
  extraMounts:
  - hostPath: ${script_dir}/data
    containerPath: /data
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${reg_name}:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
EOF