ensure_kubeconfig_path () {
   if [ ! -f "$1" ]; then
       echo "kubeconfig [$1] is missing!"
       echo "Please setup your kubeconfig to allow access to your Kubernetes cluster."
       exit 1
   fi
}

ensure_kubeconfig () {
  if [[ -z "${KUBECONFIG}" ]]; then
    ensure_kubeconfig_path "${HOME}/.kube/config"
  else
    echo "KUBECONFIG is setup with ${KUBECONFIG}"
    echo
    ensure_kubeconfig_path "${KUBECONFIG}"
  fi
}

ensure_kubectl(){
  kubectl_cmd='kubectl'
  if ! kubectl_loc="$(type -p "$kubectl_cmd")" || [[ -z $kubectl_loc ]]; then
    echo "kubectl is missing!"
    echo "Please install kubectl and try again, https://kubernetes.io/docs/tasks/tools/#kubectl"
    exit 1
  fi
}
