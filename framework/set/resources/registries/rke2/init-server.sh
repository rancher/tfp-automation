#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
RANCHER_IMAGE=$6
RANCHER_TAG_VERSION=$7
REGISTRY=$8
STAGING_RANCHER_AGENT_IMAGE=${9:-}
PRIME_RANCHER_AGENT_IMAGE=${10:-}

set -e

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

sudo tee -a /etc/rancher/rke2/registries.yaml > /dev/null << EOF
mirrors:
  docker.io:
    endpoint:
      - "https://${REGISTRY}"
configs:
  "${REGISTRY}":
    tls:
      insecure_skip_verify: true
EOF

curl -sfL https://get.rke2.io --output install.sh
sudo chmod +x install.sh
sudo INSTALL_RKE2_VERSION=${K8S_VERSION} INSTALL_RKE2_TYPE='server' ./install.sh

sudo systemctl enable rke2-server
sudo systemctl start rke2-server

sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
  "insecure-registries" : [ "${REGISTRY}" ]
}
EOF

sudo systemctl restart docker && sudo systemctl daemon-reload

if [ -n "$STAGING_RANCHER_AGENT_IMAGE" ] || [ -n "$PRIME_RANCHER_AGENT_IMAGE" ]; then
  sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}

  if [ -n "$STAGING_RANCHER_AGENT_IMAGE" ]; then
    sudo docker pull ${REGISTRY}/${STAGING_RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}
  fi

  if [ -n "$PRIME_RANCHER_AGENT_IMAGE" ]; then
    sudo docker pull ${REGISTRY}/${PRIME_RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}
  fi

  sudo systemctl restart rke2-server
fi

sudo mkdir -p /home/${USER}/.kube
sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube