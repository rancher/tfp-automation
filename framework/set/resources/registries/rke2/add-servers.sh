#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
RANCHER_IMAGE=$6
RANCHER_TAG_VERSION=$7
REGISTRY=$8
REGISTRY_USERNAME=$9
REGISTRY_PASSWORD=${10}
RANCHER_AGENT_IMAGE=${11}

set -e

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "server: https://${RKE2_SERVER_IP}:9345
token: ${RKE2_TOKEN}
system-default-registry: ${REGISTRY}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null


sudo tee /etc/rancher/rke2/registries.yaml > /dev/null << EOF
mirrors:
  "docker.io":
    endpoint:
      - "https://${REGISTRY}"
    rewrite:
      "^rancher/(.*)": "${REGISTRY}/rancher/\$1"
configs:
  "${REGISTRY}":
    tls:
      insecure_skip_verify: true
EOF

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

curl -fsSL --max-time 30 -o /home/${USER}/rke2.linux-${ARCH}.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-${ARCH}.tar.gz
curl -fsSL --max-time 30 -o /home/${USER}/rke2-images.linux-${ARCH}.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-${ARCH}.tar.zst
curl -fsSL --max-time 30 -o /home/${USER}/sha256sum-${ARCH}.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-${ARCH}.txt
curl -fsSL --max-time 30 -o /home/${USER}/install.sh https://get.rke2.io

echo "Validating checksum for rke2-images.linux-${ARCH}.tar.zst"
ZIP_NAME="rke2-images.linux-${ARCH}.tar.zst"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" /home/${USER}/sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM  /home/${USER}/rke2-images.linux-${ARCH}.tar.zst" | sha256sum -c -

sudo chmod +x /home/${USER}/install.sh
sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh /home/${USER}/install.sh
sudo systemctl enable rke2-server
sudo systemctl start rke2-server

sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
  "insecure-registries" : [ "${REGISTRY}" ]
}
EOF

sudo systemctl restart docker && sudo systemctl daemon-reload

if [ -n "$RANCHER_AGENT_IMAGE" ]; then
  sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}
  sudo docker pull ${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}
  sudo systemctl restart rke2-server
fi