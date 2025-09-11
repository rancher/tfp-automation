#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_TOKEN=$5
REGISTRY_USERNAME=$6
REGISTRY_PASSWORD=$7
CLUSTER_CIDR=$8
SERVICE_CIDR=$9

set -e

sudo hostnamectl set-hostname ${K3S_SERVER_IP}

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Installing kubectl"
sudo curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl"
sudo chmod +x kubectl

sudo mv kubectl /usr/local/bin/

sudo mkdir -p /etc/rancher/k3s
sudo touch /etc/rancher/k3s/registries.yaml

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

curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} sh -s - server --cluster-init --cluster-cidr=${CLUSTER_CIDR} --service-cidr=${SERVICE_CIDR}

sudo mkdir -p /home/${USER}/.kube
sudo chown ${USER}:${GROUP} /etc/rancher/k3s/k3s.yaml
sudo cp /etc/rancher/k3s/k3s.yaml /home/${USER}/.kube/config
sudo chown ${USER}:${GROUP} /home/${USER}/.kube/config