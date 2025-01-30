#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
BASTION=$6
PORT="3128"

set -e

cat <<EOF | sudo tee /etc/default/rke2-server > /dev/null
HTTP_PROXY=http://${BASTION}:${PORT}
HTTPS_PROXY=http://${BASTION}:${PORT}
NO_PROXY=localhost,127.0.0.0/8,10.0.0/8,cattle-system.svc,172.16.0.0/12,192.168.0.0/16,.svc,.cluster.local
CONTAINERD_HTTP_PROXY=http://${BASTION}:${PORT}
CONTAINERD_HTTPS_PROXY=http://${BASTION}:${PORT}
CONTAINERD_NO_PROXY=localhost,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.169.254,.svc,.cluster.local,cattle-system.svc
http_proxy=http://${BASTION}:${PORT}
https_proxy=http://${BASTION}:${PORT}
EOF

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "server: https://${RKE2_SERVER_IP}:9345
token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh
sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' ./install.sh

sudo systemctl enable rke2-server
sudo systemctl start rke2-server