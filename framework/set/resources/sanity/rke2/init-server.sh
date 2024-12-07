#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5

set -e

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh
sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' ./install.sh

sudo systemctl enable rke2-server
sudo systemctl start rke2-server

sudo mkdir -p /home/${USER}/.kube
sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube