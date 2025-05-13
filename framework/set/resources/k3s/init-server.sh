#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_TOKEN=$5

set -e

sudo hostnamectl set-hostname ${K3S_SERVER_IP}

curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} sh -s - server --cluster-init

sudo mkdir -p /home/${USER}/.kube
sudo chown ${USER}:${GROUP} /etc/rancher/k3s/k3s.yaml
sudo cp /etc/rancher/k3s/k3s.yaml /home/${USER}/.kube/config
sudo chown ${USER}:${GROUP} /home/${USER}/.kube/config