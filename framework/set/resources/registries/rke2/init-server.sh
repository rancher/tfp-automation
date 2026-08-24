#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
RANCHER_IMAGE=$6
RANCHER_TAG_VERSION=$7
REGISTRY=$8
REPO=$9
RANCHER_CHART_REPO=${10}
REGISTRY_USERNAME=${11}
REGISTRY_PASSWORD=${12}
DOCKERHUB_USER=${13}
DOCKERHUB_PASS=${14}
RANCHER_AGENT_IMAGE=${15}
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

if [[ $REPO == "prime-release" ]]; then
  . /etc/os-release

  [[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt -y install yq
  [[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install yq -y
  [[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install  -y yq

  HEAD_SERIES=""
  if [[ $RANCHER_TAG_VERSION =~ ^v([0-9]+\.[0-9]+)-head$ ]]; then
      HEAD_SERIES="${BASH_REMATCH[1]}"
  fi

  if [[ -z "$HEAD_SERIES" && $RANCHER_TAG_VERSION == "head" ]]; then
    HEAD_SERIES="${RANCHER_HEAD_SERIES:-}"
    if [[ -z "$HEAD_SERIES" ]]; then
      HEAD_SERIES=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher[].version' | sed -nE 's/^([0-9]+\.[0-9]+)\.[0-9]+.*-head$/\1/p' | sort -Vu | tail -n 1)
    fi
  fi

  if [[ $RANCHER_TAG_VERSION == "head" && -z "$HEAD_SERIES" ]]; then
      LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher | map(select(.version | test("^"))) | sort_by(.created) | .[-1].version')
  fi

  if [[ -n "$HEAD_SERIES" ]]; then
      SERIES_FILTER="^${HEAD_SERIES//./\\\\.}"
      LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r ".entries.rancher | map(select(.version | test(\"${SERIES_FILTER}\"))) | sort_by(.created) | .[-1].version")
  fi
fi

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "token: ${RKE2_TOKEN}
system-default-registry: ${REGISTRY}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

if [ -z "${REGISTRY_USERNAME}" ] || [ -z "${REGISTRY_PASSWORD}" ]; then
  sudo tee /etc/rancher/rke2/registries.yaml > /dev/null << EOF
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
else
  sudo tee /etc/rancher/rke2/registries.yaml > /dev/null << EOF
mirrors:
  "docker.io":
    endpoint:
      - "https://${REGISTRY}"
    rewrite:
      "^rancher/(.*)": "${REGISTRY}/rancher/\$1"
configs:
  "${REGISTRY}":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
EOF
fi

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

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

retryCmd curl -sfL https://get.rke2.io --output /home/${USER}/install.sh
sudo chmod +x /home/${USER}/install.sh

retryCmd sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh /home/${USER}/install.sh
retryCmd sudo systemctl enable rke2-server
retryCmd sudo systemctl start rke2-server

sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
  "insecure-registries" : [ "${REGISTRY}" ]
}
EOF

sudo systemctl restart docker && sudo systemctl daemon-reload

if [ -n "${REGISTRY_USERNAME}" ] && [ -n "${REGISTRY_PASSWORD}" ]; then
  sudo docker login https://registry-1.docker.io -u "${DOCKERHUB_USER}" -p "${DOCKERHUB_PASS}"
  sudo docker login https://${REGISTRY} -u "${REGISTRY_USERNAME}" -p "${REGISTRY_PASSWORD}"
fi

if [ -n "$RANCHER_AGENT_IMAGE" ]; then
  if [ "$REPO" == "prime-release" ]; then
    RANCHER_TAG_VERSION="v${LATEST_CHART_VERSION}"
  fi

  sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}
  sudo docker pull ${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}
  sudo systemctl restart rke2-server
fi

sudo mkdir -p /home/${USER}/.kube
sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube