#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_TOKEN=$5
REGISTRY_USERNAME=$6
REGISTRY_PASSWORD=$7
CLUSTER_CIDR=$8
SERVICE_CIDR=$9
MAX_CMD_RETRIES=20
CMD_RETRY_INTERVAL_SECONDS=10

set -e

retryCmd() {
  local attempt=1
  local rc=0

  while [ "$attempt" -le "$MAX_CMD_RETRIES" ]; do
    if "$@"; then
      return 0
    else
      rc=$?
    fi

    if [ "$attempt" -eq "$MAX_CMD_RETRIES" ]; then
      echo "Command failed after ${MAX_CMD_RETRIES} attempts (exit ${rc}): $*" >&2
      return "$rc"
    fi

    echo "Command failed on attempt ${attempt}/${MAX_CMD_RETRIES} (exit ${rc}), retrying in ${CMD_RETRY_INTERVAL_SECONDS}s: $*" >&2
    sleep "$CMD_RETRY_INTERVAL_SECONDS"
    attempt=$((attempt + 1))
  done

  return "$rc"
}

sudo hostnamectl set-hostname ${K3S_SERVER_IP}

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Installing kubectl"
KUBECTL_VERSION="v1.36.0"
retryCmd curl -fsSL --max-time 30 -o kubectl https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl
retryCmd curl -fsSL --max-time 30 -o kubectl.sha256 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256
echo "$(cat kubectl.sha256) kubectl" | sha256sum -c
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

sudo mkdir -p /etc/rancher/k3s
sudo touch /etc/rancher/k3s/registries.yaml

echo "mirrors:
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
      password: "${REGISTRY_PASSWORD}"" | sudo tee -a /etc/rancher/k3s/registries.yaml > /dev/null

retryCmd curl -fsSL --max-time 30 https://get.k3s.io | INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} sh -s - server --cluster-init --cluster-cidr=${CLUSTER_CIDR} --service-cidr=${SERVICE_CIDR}

sudo mkdir -p /home/${USER}/.kube
sudo chown ${USER}:${GROUP} /etc/rancher/k3s/k3s.yaml
sudo cp /etc/rancher/k3s/k3s.yaml /home/${USER}/.kube/config
sudo chown ${USER}:${GROUP} /home/${USER}/.kube/config