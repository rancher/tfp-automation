#!/bin/bash

K8S_VERSION=$1
RKE2_SERVER_ONE_IP=$2
RKE2_SERVER_TWO_IP=$3
RKE2_SERVER_THREE_IP=$4
USER=$5
PEM_FILE=$6

set -e

base64 -d <<< $PEM_FILE > /home/$USER/airgap.pem
PEM=/home/$USER/airgap.pem
chmod 600 $PEM

wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2.linux-amd64.tar.gz
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2.linux-arm64.tar.gz
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2-images.linux-amd64.tar.zst
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2-images.linux-arm64.tar.zst
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/sha256sum-amd64.txt
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/sha256sum-arm64.txt

curl -sfL https://get.rke2.io --output install.sh
chmod +x install.sh

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    KUBECTL_ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    KUBECTL_ARCH="arm64"
fi

echo "Installing kubectl"
sudo curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${KUBECTL_ARCH}/kubectl"
sudo chmod +x kubectl
sudo mv kubectl /usr/local/bin/

echo "Copying files to RKE2 server one"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null /usr/local/bin/kubectl ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/

echo "Copying files to RKE2 server two"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/

echo "Copying files to RKE2 server three"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/