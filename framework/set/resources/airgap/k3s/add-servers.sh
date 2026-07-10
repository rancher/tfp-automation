#!/bin/bash

USER=$1
GROUP=$2
VPC_IP=$3
K8S_VERSION=$4
K3S_SERVER_ONE_IP=$5
K3S_NEW_SERVER_IP=$6
K3S_TOKEN=$7
REGISTRY=$8
REGISTRY_USERNAME=$9
REGISTRY_PASSWORD=${10}
RANCHER_IMAGE=${11}
RANCHER_TAG_VERSION=${12}
RANCHER_AGENT_IMAGE=${13}
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
       export VPC_IP=${VPC_IP}; \
       export K3S_SERVER_ONE_IP=${K3S_SERVER_ONE_IP}; \
       export K3S_NEW_SERVER_IP=${K3S_NEW_SERVER_IP}; \
       export K3S_TOKEN=${K3S_TOKEN}; \
       export REGISTRY=${REGISTRY}; \
       export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
       export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; $cmd"; then
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
token: ${K3S_TOKEN}
system-default-registry: ${REGISTRY}
tls-san:
  - ${K3S_SERVER_ONE_IP}
EOF
}

setup_registry() {
  sudo tee /etc/rancher/k3s/registries.yaml > /dev/null << EOF
mirrors:
  "docker.io":
    endpoint:
      - "https://${REGISTRY}"
    rewrite:
      "^rancher/(.*)": "${REGISTRY}/rancher/\$1"
configs:
  "${REGISTRY}":
    tls:
      insecure_skip_verify: true
EOF
}

setup_docker_daemon() {
  sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
  "insecure-registries" : [ "${REGISTRY}" ]
}
EOF
}

setup_networking() {
  sudo systemctl disable systemd-resolved; sudo systemctl stop systemd-resolved
  sleep 5
  sudo sed -i.bak "s/^nameserver .*/nameserver ${VPC_IP}/" /etc/resolv.conf
  sudo sed -i.bak "/^options /d" /etc/resolv.conf
  echo "options edns0" | sudo tee -a /etc/resolv.conf

  sudo tee /etc/NetworkManager/conf.d/90-dns-none.conf > /dev/null << EOF
[main]
dns=none
EOF
}

run_ssh "${K3S_NEW_SERVER_IP}" "sudo mv /home/${USER}/k3s /usr/local/bin/"

configFunction=$(declare -f setup_config)
run_ssh "${K3S_NEW_SERVER_IP}" "${configFunction}; setup_config"

setupRegistryFunction=$(declare -f setup_registry)
run_ssh "${K3S_NEW_SERVER_IP}" "${setupRegistryFunction}; setup_registry"

run_ssh "${K3S_NEW_SERVER_IP}" "sudo INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_EXEC=\"server --server https://${K3S_SERVER_ONE_IP}:6443\" INSTALL_K3S_SKIP_DOWNLOAD=true sh install.sh"

setupDaemonFunction=$(declare -f setup_docker_daemon)
run_ssh "${K3S_NEW_SERVER_IP}" "${setupDaemonFunction}; setup_docker_daemon"
run_ssh "${K3S_NEW_SERVER_IP}" "sudo systemctl restart docker && sudo systemctl daemon-reload"

setupNetworkingFunction=$(declare -f setup_networking)
run_ssh "${K3S_NEW_SERVER_IP}" "${setupNetworkingFunction}; setup_networking"

if [ -n "$RANCHER_AGENT_IMAGE" ]; then
  run_ssh "${K3S_NEW_SERVER_IP}" "sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}"
  run_ssh "${K3S_NEW_SERVER_IP}" "sudo docker pull ${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}"
  run_ssh "${K3S_NEW_SERVER_IP}" "sudo systemctl restart k3s"
fi

kubectl get nodes