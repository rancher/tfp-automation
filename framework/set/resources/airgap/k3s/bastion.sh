#!/bin/bash

K8S_VERSION=$1
K3S_SERVER_ONE_IP=$2
K3S_SERVER_TWO_IP=$3
K3S_SERVER_THREE_IP=$4
USER=$5
PEM_FILE=$6
REPO=$7
RANCHER_TAG_VERSION=$8
RANCHER_CHART_REPO=$9
REQUIRED_PACKAGES=("curl")
MISSING_PACKAGES=()

set -e

for pkg in "${REQUIRED_PACKAGES[@]}"; do
    if ! command -v "$pkg" &> /dev/null; then
        MISSING_PACKAGES+=("$pkg")
    fi
done

if [ "${#MISSING_PACKAGES[@]}" -gt 0 ]; then
    echo "Error: missing required packages: ${MISSING_PACKAGES[*]}"
    exit 1
fi

base64 -d <<< $PEM_FILE > /home/$USER/airgap.pem
PEM=/home/$USER/airgap.pem
chmod 600 $PEM

curl -fsSL --max-time 30 -o k3s https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s
curl -fsSL --max-time 30 -o k3s-images.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-images.txt
curl -fsSL --max-time 30 -o k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-amd64.tar.gz
curl -fsSL --max-time 30 -o k3s-airgap-images-arm64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-arm64.tar.gz
curl -fsSL --max-time 30 -o sha256sum-amd64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-amd64.txt
curl -fsSL --max-time 30 -o sha256sum-arm64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-arm64.txt
curl -fsSL --max-time 30 -o install.sh https://get.k3s.io

if [[ $REPO == "prime-release" ]]; then
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
        RANCHER_TAG_VERSION="${LATEST_CHART_VERSION}"
    fi

    if [[ -n "$HEAD_SERIES" ]]; then
        SERIES_FILTER="^${HEAD_SERIES//./\\\\.}"
        LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r ".entries.rancher | map(select(.version | test(\"${SERIES_FILTER}\"))) | sort_by(.created) | .[-1].version")
        RANCHER_TAG_VERSION="${LATEST_CHART_VERSION}"
    fi

    printf '%s\n' "${RANCHER_TAG_VERSION}" > "/home/${USER}/rancher-tag-version.txt"
fi

chmod +x k3s
chmod +x install.sh

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Validating checksum for k3s-airgap-images-${ARCH}.tar.gz"
ZIP_NAME="k3s-airgap-images-${ARCH}.tar.gz"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM k3s-airgap-images-${ARCH}.tar.gz" | sha256sum -c -

echo "Copying files to K3S server one"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-amd64.tar.gz ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-arm64.tar.gz ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/

echo "Copying files to K3S server two"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-amd64.tar.gz ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-arm64.tar.gz ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/

echo "Copying files to K3S server three"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-amd64.tar.gz ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-arm64.tar.gz ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/

if [[ $REPO == "prime-release" ]]; then
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "/home/${USER}/rancher-tag-version.txt" ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/rancher-tag-version.txt
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "/home/${USER}/rancher-tag-version.txt" ${USER}@${K3S_SERVER_TWO_IP}:/home/${USER}/rancher-tag-version.txt
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "/home/${USER}/rancher-tag-version.txt" ${USER}@${K3S_SERVER_THREE_IP}:/home/${USER}/rancher-tag-version.txt
fi

mkdir -p ~/.kube