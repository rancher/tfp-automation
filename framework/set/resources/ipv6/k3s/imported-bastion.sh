#!/bin/bash

K8S_VERSION=$1
K3S_SERVER_ONE_IP=$2

ARGS=("$@")
ARG_COUNT=$#
USER=${ARGS[$((ARG_COUNT - 2))]}
PEM_FILE=${ARGS[$((ARG_COUNT - 1))]}
K3S_SERVER_ADDITIONAL_IPS=("${ARGS[@]:2:$((ARG_COUNT - 4))}")

set -e

base64 -d <<< "$PEM_FILE" > /home/$USER/airgap.pem
PEM=/home/$USER/airgap.pem
chmod 600 $PEM

curl -fsSL --max-time 30 -o k3s https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s
curl -fsSL --max-time 30 -o k3s-images.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-images.txt
curl -fsSL --max-time 30 -o k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-amd64.tar.gz
curl -fsSL --max-time 30 -o k3s-airgap-images-arm64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/k3s-airgap-images-arm64.tar.gz
curl -fsSL --max-time 30 -o sha256sum-amd64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-amd64.txt
curl -fsSL --max-time 30 -o sha256sum-arm64.txt https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}/sha256sum-arm64.txt
curl -fsSL --max-time 30 -o install.sh https://get.k3s.io

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