#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_ONE_PUBLIC_IP=$4
K3S_SERVER_ONE_PRIVATE_IP=$5
K3S_NEW_SERVER_PRIVATE_IP=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
K3S_TOKEN=$9
CLUSTER_CIDR=${10}
SERVICE_CIDR=${11}
PEM_FILE=/home/$USER/airgap.pem
MAX_SSH_RETRIES=20
SSH_RETRY_INTERVAL_SECONDS=10

set -e

run_ssh() {
  local server="$1"
  local cmd="$2"
  local attempt=1
  local rc=0

  while [ "$attempt" -le "$MAX_SSH_RETRIES" ]; do
    if ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes -o ConnectTimeout=20 -o ServerAliveInterval=30 -o ServerAliveCountMax=10 -i "$PEM_FILE" "$USER@$server" \
      "export USER=${USER}; \
       export GROUP=${GROUP}; \
       export K3S_SERVER_ONE_PUBLIC_IP=${K3S_SERVER_ONE_PUBLIC_IP}; \
       export K3S_SERVER_ONE_PRIVATE_IP=${K3S_SERVER_ONE_PRIVATE_IP}; \
       export K3S_NEW_SERVER_PRIVATE_IP=${K3S_NEW_SERVER_PRIVATE_IP}; \
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

setup_config() {
    sudo mkdir -p /etc/rancher/k3s
    sudo tee /etc/rancher/k3s/config.yaml > /dev/null << EOF
write-kubeconfig-mode: 644
token: ${K3S_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
flannel-ipv6-masq: true
tls-san:
  - ${K3S_SERVER_ONE_PRIVATE_IP}
EOF
}

setup_registry() {
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

run_ssh "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo mv /home/${USER}/k3s /usr/local/bin/"

configFunction=$(declare -f setup_config)
run_ssh "${K3S_NEW_SERVER_PRIVATE_IP}" "${configFunction}; setup_config"

registryFunction=$(declare -f setup_registry)
run_ssh "${K3S_NEW_SERVER_PRIVATE_IP}" "${registryFunction}; setup_registry"

run_ssh "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_EXEC=\"server --server https://[${K3S_SERVER_ONE_PUBLIC_IP}]:6443\" INSTALL_K3S_SKIP_DOWNLOAD=true sh install.sh"