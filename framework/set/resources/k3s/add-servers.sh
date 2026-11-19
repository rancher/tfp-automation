#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_NEW_SERVER_IP=$5
K3S_TOKEN=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
ETCD_S3_ACCESS_KEY=${9}
ETCD_S3_SECRET_KEY=${10}
ETCD_S3_ENDPOINT=${11}
ETCD_S3_BUCKET=${12}
ETCD_S3_REGION=${13}
MAX_CMD_RETRIES=20
CMD_RETRY_INTERVAL_SECONDS=10
REQUIRED_PACKAGES=("curl")
MISSING_PACKAGES=()

set -e

retry_cmd() {
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

for pkg in "${REQUIRED_PACKAGES[@]}"; do
  if ! command -v "$pkg" &> /dev/null; then
    MISSING_PACKAGES+=("$pkg")
  fi
done

if [ "${#MISSING_PACKAGES[@]}" -gt 0 ]; then
  echo "Error: missing required packages: ${MISSING_PACKAGES[*]}"
  exit 1
fi

sudo hostnamectl set-hostname ${K3S_NEW_SERVER_IP}
sudo mkdir -p /etc/rancher/k3s

if [ -n "${ETCD_S3_ACCESS_KEY}" ]; then
  echo "token: ${K3S_TOKEN}
secrets-encryption: true
etcd-s3: true
etcd-s3-access-key: ${ETCD_S3_ACCESS_KEY}
etcd-s3-secret-key: ${ETCD_S3_SECRET_KEY}
etcd-s3-endpoint: ${ETCD_S3_ENDPOINT}
etcd-s3-bucket: ${ETCD_S3_BUCKET}
etcd-s3-region: ${ETCD_S3_REGION}
etcd-s3-insecure: true
tls-san:
  - ${K3S_SERVER_IP}" | sudo tee /etc/rancher/k3s/config.yaml > /dev/null
else
  echo "server: https://${K3S_SERVER_IP}:6443
token: ${K3S_TOKEN}
secrets-encryption: true
tls-san:
  - ${K3S_SERVER_IP}" | sudo tee /etc/rancher/k3s/config.yaml > /dev/null
fi

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

retry_cmd curl -fsSL --max-time 120 -o install.sh https://get.k3s.io
chmod +x install.sh
retry_cmd sudo INSTALL_K3S_VERSION=${K8S_VERSION} INSTALL_K3S_EXEC=server sh install.sh