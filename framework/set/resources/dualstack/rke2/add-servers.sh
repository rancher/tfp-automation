#!/bin/bash

USER=$1
K8S_VERSION=$2
RKE2_SERVER_ONE_IP=$3
RKE2_NEW_SERVER_IP=$4
RKE2_TOKEN=$5
CNI=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
CLUSTER_CIDR=$9
SERVICE_CIDR=${10}

set -e

sudo hostnamectl set-hostname ${RKE2_NEW_SERVER_IP}

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-${ARCH}.tar.gz
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-${ARCH}.tar.zst
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-${ARCH}.txt

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "server: https://${RKE2_SERVER_ONE_IP}:9345
cni: ${CNI}
token: ${RKE2_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
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
      password: "${REGISTRY_PASSWORD}"" | sudo tee -a /etc/rancher/rke2/registries.yaml > /dev/null

curl -sfL https://get.rke2.io --output install.sh
chmod +x install.sh

sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh
sudo systemctl enable rke2-server
sudo systemctl start rke2-server