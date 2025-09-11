#!/bin/bash

K8S_VERSION=$1
K3S_SERVER_ONE_IP=$2
K3S_SERVER_TWO_IP=$3
K3S_SERVER_THREE_IP=$4
USER=$5
PEM_FILE=$6

set -e

base64 -d <<< $PEM_FILE > /home/$USER/airgap.pem
PEM=/home/$USER/airgap.pem
chmod 600 $PEM

wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/k3s
wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/k3s-images.txt
wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/k3s-airgap-images-amd64.tar.gz
wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/k3s-airgap-images-arm64.tar.gz
wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/sha256sum-amd64.txt
wget https://github.com/k3s-io/k3s/releases/download/${K8S_VERSION}+k3s1/sha256sum-arm64.txt

curl -sfL https://get.k3s.io --output install.sh
chmod +x k3s
chmod +x install.sh

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Installing kubectl"
sudo curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl"
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

sudo mv kubectl /usr/local/bin/