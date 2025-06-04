#!/bin/bash

K8S_VERSION=$1
RKE2_SERVER_IP=$2
RKE2_NEW_SERVER_IP=$3
HOSTNAME=$4
RKE2_TOKEN=$5
CNI=$6
CLUSTER_CIDR=${7}
SERVICE_CIDR=${8}

set -e

sudo hostnamectl set-hostname ${RKE2_NEW_SERVER_IP}

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

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
  - ${HOSTNAME}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
else
  echo "server: https://${RKE2_SERVER_IP}:9345
cni: ${CNI}
token: ${RKE2_TOKEN}
tls-san:
  - ${HOSTNAME}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
fi

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh

sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' sh ./install.sh

sudo systemctl enable rke2-server
sudo systemctl start rke2-server