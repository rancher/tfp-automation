#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_NEW_SERVER_IP=$5
K3S_TOKEN=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8

set -e

sudo hostnamectl set-hostname ${K3S_NEW_SERVER_IP}

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

curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} sh -s - server --server https://${K3S_SERVER_IP}:6443