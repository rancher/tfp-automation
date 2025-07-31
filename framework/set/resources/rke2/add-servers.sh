#!/bin/bash

K8S_VERSION=$1
RKE2_SERVER_ONE_IP=$2
RKE2_NEW_SERVER_IP=$3
RKE2_TOKEN=$4
CNI=$5
CLUSTER_CIDR=${6}
SERVICE_CIDR=${7}

set -e

sudo hostnamectl set-hostname ${RKE2_NEW_SERVER_IP}

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

createConfigYAML() {
  if [ -n "${CLUSTER_CIDR}" ]; then
    echo "server: https://${RKE2_SERVER_IP}:9345
write-kubeconfig-mode: 644
node-ip: ${RKE2_NEW_SERVER_IP}
node-external-ip: ${RKE2_NEW_SERVER_IP}
cni: ${CNI}
token: ${RKE2_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
tls-san:
  - ${RKE2_SERVER_ONE_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
  else
    echo "server: https://${RKE2_SERVER_ONE_IP}:9345
cni: ${CNI}
token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_ONE_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
  fi
}

createConfigYAML

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh
sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' sh ./install.sh

sudo systemctl enable rke2-server

if ! sudo systemctl start rke2-server; then
  sudo /usr/local/bin/rke2-killall.sh
  sudo /usr/local/bin/rke2-uninstall.sh
  sudo rm -rf /var/lib/rancher/rke2
  sudo rm -rf /etc/rancher/rke2

  sudo mkdir -p /etc/rancher/rke2
  sudo touch /etc/rancher/rke2/config.yaml

  createConfigYAML

  sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' ./install.sh
fi