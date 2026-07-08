#!/bin/bash

USER=$1
GROUP=$2
K3S_SERVER_PUBLIC_IP=$3
K3S_SERVER_PRIVATE_IP=$4
REGISTRY_USERNAME=$5
REGISTRY_PASSWORD=$6
K3S_TOKEN=$7
CLUSTER_CIDR=$8
SERVICE_CIDR=$9
PEM_FILE=/home/$USER/airgap.pem
MAX_SSH_RETRIES=20
SSH_RETRY_INTERVAL_SECONDS=10

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  local attempt=1
  local rc=0

  while [ "$attempt" -le "$MAX_SSH_RETRIES" ]; do
    if ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes -o ConnectTimeout=20 -o ServerAliveInterval=30 -o ServerAliveCountMax=10 -i "$PEM_FILE" "$USER@$server" \
      "export USER=${USER}; \
       export GROUP=${GROUP}; \
       export K3S_SERVER_PUBLIC_IP=${K3S_SERVER_PUBLIC_IP}; \
       export K3S_SERVER_PRIVATE_IP=${K3S_SERVER_PRIVATE_IP}; \
       export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
       export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; \
       export K3S_TOKEN=${K3S_TOKEN}; \
       export CLUSTER_CIDR=${CLUSTER_CIDR}; \
       export SERVICE_CIDR=${SERVICE_CIDR}; $cmd"; then
      return 0
    else
      rc=$?
    fi

    if [ "$attempt" -eq "$MAX_SSH_RETRIES" ]; then
      echo "SSH command failed after ${MAX_SSH_RETRIES} attempts (exit ${rc})" >&2
      return "$rc"
    fi

    echo "SSH command failed on attempt ${attempt}/${MAX_SSH_RETRIES} (exit ${rc}), retrying in ${SSH_RETRY_INTERVAL_SECONDS}s..." >&2
    sleep "$SSH_RETRY_INTERVAL_SECONDS"
    attempt=$((attempt + 1))
  done

  return "$rc"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/k3s
    sudo tee /etc/rancher/k3s/config.yaml > /dev/null << EOF
write-kubeconfig-mode: 644
token: ${K3S_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
cluster-init: true
flannel-ipv6-masq: true
tls-san:
  - ${K3S_SERVER_PRIVATE_IP}
EOF
}

setupRegistry() {
  sudo tee /etc/rancher/k3s/registries.yaml > /dev/null << EOF
mirrors:
  docker.io:
    endpoint:
      - "https://registry-1.docker.io"
configs:
  "registry-1.docker.io":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
  "docker.io":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
EOF
}

runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo mv /home/${USER}/kubectl /usr/local/bin/"
runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo mv /home/${USER}/k3s /usr/local/bin/"

configFunction=$(declare -f setupConfig)
runSSH "${K3S_SERVER_PRIVATE_IP}" "${configFunction}; setupConfig"

registryFunction=$(declare -f setupRegistry)
runSSH "${K3S_SERVER_PRIVATE_IP}" "${registryFunction}; setupRegistry"

runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_EXEC=server INSTALL_K3S_SKIP_DOWNLOAD=true sh install.sh"

runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo mkdir -p /home/${USER}/.kube"
runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo cp /etc/rancher/k3s/k3s.yaml /home/${USER}/.kube/config"
runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube"

if [ ! -s ~/.kube/config ]; then
  runSSH "${K3S_SERVER_PRIVATE_IP}" "sudo cat /home/${USER}/.kube/config" > ~/.kube/config
fi

sed -i "s|\[::1\]|[${K3S_SERVER_PUBLIC_IP}]|g" ~/.kube/config
kubectl get nodes