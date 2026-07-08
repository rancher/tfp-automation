#!/bin/bash

USER=$1
K8S_VERSION=$2
RKE2_SERVER_ONE_IP=$3
RKE2_NEW_SERVER_IP=$4
RKE2_TOKEN=$5
CNI=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
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

sudo hostnamectl set-hostname ${RKE2_NEW_SERVER_IP}

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

retryCmd curl -fsSL --max-time 30 -o rke2.linux-${ARCH}.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-${ARCH}.tar.gz
retryCmd curl -fsSL --max-time 30 -o rke2-images.linux-${ARCH}.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-${ARCH}.tar.zst
retryCmd curl -fsSL --max-time 30 -o sha256sum-${ARCH}.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-${ARCH}.txt

echo "Validating checksum for rke2.linux-${ARCH}.tar.gz"
ZIP_NAME="rke2.linux-${ARCH}.tar.gz"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM  rke2.linux-${ARCH}.tar.gz" | sha256sum -c -

if [[ "${USER}" == "root" ]]; then
  mkdir -p /home/root
  mv rke2.linux-${ARCH}.tar.gz /home/root/
  mv rke2-images.linux-${ARCH}.tar.zst /home/root/
  mv sha256sum-${ARCH}.txt /home/root/
fi

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "server: https://${RKE2_SERVER_ONE_IP}:9345
cni: ${CNI}
token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_ONE_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

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
      password: "${REGISTRY_PASSWORD}"" | sudo tee /etc/rancher/rke2/registries.yaml > /dev/null

retryCmd curl -fsSL --max-time 30 -o install.sh https://get.rke2.io
chmod +x install.sh

retryCmd sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh
retryCmd sudo systemctl enable rke2-server
retryCmd sudo systemctl start rke2-server