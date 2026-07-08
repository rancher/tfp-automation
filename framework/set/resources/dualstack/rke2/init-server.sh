#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
CNI=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
CLUSTER_CIDR=$9
SERVICE_CIDR=${10}
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

sudo hostnamectl set-hostname ${RKE2_SERVER_IP}

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

retryCmd curl -fsSL --max-time 30 -o /home/${USER}/rke2.linux-${ARCH}.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-${ARCH}.tar.gz
retryCmd curl -fsSL --max-time 30 -o /home/${USER}/rke2-images.linux-${ARCH}.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-${ARCH}.tar.zst
retryCmd curl -fsSL --max-time 30 -o /home/${USER}/sha256sum-${ARCH}.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-${ARCH}.txt

echo "Validating checksum for rke2-images.linux-${ARCH}.tar.zst"
ZIP_NAME="rke2-images.linux-${ARCH}.tar.zst"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" /home/${USER}/sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM  /home/${USER}/rke2-images.linux-${ARCH}.tar.zst" | sha256sum -c -

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "cni: ${CNI}
write-kubeconfig-mode: 644
token: ${RKE2_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

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
      password: "${REGISTRY_PASSWORD}"" | sudo tee -a /etc/rancher/rke2/registries.yaml > /dev/null

retryCmd curl -sfL https://get.rke2.io --output /home/${USER}/install.sh
chmod +x /home/${USER}/install.sh

retryCmd sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh /home/${USER}/install.sh
retryCmd sudo systemctl enable rke2-server
retryCmd sudo systemctl start rke2-server

if [[ "${USER}" == "root" ]]; then
  sudo mkdir -p /root/.kube
  sudo cp /etc/rancher/rke2/rke2.yaml /root/.kube/config
else
  sudo mkdir -p /home/${USER}/.kube
  sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
  sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube
fi