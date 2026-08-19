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

set -e

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

echo "Installing kubectl"
KUBECTL_VERSION="v1.36.0"
curl -fsSL --max-time 30 -o kubectl https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl
curl -fsSL --max-time 30 -o kubectl.sha256 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256
echo "$(cat kubectl.sha256) kubectl" | sha256sum -c
sudo chmod +x kubectl

echo "Copying files to K3S server one"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null kubectl ${USER}@${K3S_SERVER_ONE_IP}:/home/${USER}/
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
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
rm kubectl