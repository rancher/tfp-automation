#!/bin/bash

K8S_VERSION=$1
K3S_SERVER_ONE_IP=$2
ARGS=("$@")
ARG_COUNT=$#
USER=${ARGS[$((ARG_COUNT - 2))]}
PEM_FILE=${ARGS[$((ARG_COUNT - 1))]}
K3S_SERVER_ADDITIONAL_IPS=("${ARGS[@]:2:$((ARG_COUNT - 4))}")
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

base64 -d <<< "$PEM_FILE" > /home/$USER/airgap.pem
PEM=/home/$USER/airgap.pem
chmod 600 $PEM

curl -fsSL --max-time 120 -o k3s https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s
curl -fsSL --max-time 120 -o k3s-images.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-images.txt
curl -fsSL --max-time 120 -o k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-amd64.tar.gz
curl -fsSL --max-time 120 -o k3s-airgap-images-arm64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-arm64.tar.gz
curl -fsSL --max-time 120 -o sha256sum-amd64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-amd64.txt
curl -fsSL --max-time 120 -o sha256sum-arm64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-arm64.txt
curl -fsSL --max-time 120 -o install.sh https://get.k3s.io

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

for ip in "${K3S_SERVER_ADDITIONAL_IPS[@]}"; do
    echo "Copying files to K3S server at $ip"
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null airgap.pem ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-amd64.tar.gz ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null k3s-airgap-images-arm64.tar.gz ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${ip}:/home/${USER}/
    sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${ip}:/home/${USER}/
done