#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
CNI=$6
CLUSTER_CIDR=${7}
SERVICE_CIDR=${8}

set -e

sudo hostnamectl set-hostname ${RKE2_SERVER_IP}

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

createConfigYAML() {
  if [ -n "${CLUSTER_CIDR}" ]; then
    echo "cni: ${CNI}
write-kubeconfig-mode: 644
node-ip: ${RKE2_SERVER_IP}
node-external-ip: ${RKE2_SERVER_IP}
token: ${RKE2_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
  else
    echo "cni: ${CNI}
token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null
  fi
}

createConfigYAML

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh
sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' ./install.sh

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

if [[ "${USER}" == "root" ]]; then
  sudo mkdir -p /root/.kube
  sudo cp /etc/rancher/rke2/rke2.yaml /root/.kube/config
else
  sudo mkdir -p /home/${USER}/.kube
  sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
  sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube
fi