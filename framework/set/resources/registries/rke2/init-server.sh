#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_IP=$4
RKE2_TOKEN=$5
RANCHER_IMAGE=$6
RANCHER_TAG_VERSION=$7
REGISTRY=$8
REGISTRY_USERNAME=${9}
REGISTRY_PASSWORD=${10}
DOCKERHUB_USER=${11}
DOCKERHUB_PASS=${12}
RANCHER_AGENT_IMAGE=${13}

set -e

sudo mkdir -p /etc/rancher/rke2
sudo touch /etc/rancher/rke2/config.yaml

echo "token: ${RKE2_TOKEN}
system-default-registry: ${REGISTRY}
tls-san:
  - ${RKE2_SERVER_IP}" | sudo tee /etc/rancher/rke2/config.yaml > /dev/null

if [ -z "${REGISTRY_USERNAME}" ] || [ -z "${REGISTRY_PASSWORD}" ]; then
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
else
  sudo tee /etc/rancher/rke2/registries.yaml > /dev/null << EOF
mirrors:
  "docker.io":
    endpoint:
      - "https://${REGISTRY}"
    rewrite:
      "^rancher/(.*)": "${REGISTRY}/rancher/\$1"
configs:
  "${REGISTRY}":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
EOF
fi

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

curl -fsSL --max-time 30 -o /home/${USER}/rke2.linux-${ARCH}.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-${ARCH}.tar.gz
curl -fsSL --max-time 30 -o /home/${USER}/rke2-images.linux-${ARCH}.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-${ARCH}.tar.zst
curl -fsSL --max-time 30 -o /home/${USER}/sha256sum-${ARCH}.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-${ARCH}.txt

echo "Validating checksum for rke2-images.linux-${ARCH}.tar.zst"
ZIP_NAME="rke2-images.linux-${ARCH}.tar.zst"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" /home/${USER}/sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM  /home/${USER}/rke2-images.linux-${ARCH}.tar.zst" | sha256sum -c -

curl -sfL https://get.rke2.io --output /home/${USER}/install.sh
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

if [ -n "${REGISTRY_USERNAME}" ] && [ -n "${REGISTRY_PASSWORD}" ]; then
  sudo docker login https://registry-1.docker.io -u "${DOCKERHUB_USER}" -p "${DOCKERHUB_PASS}"
  sudo docker login https://${REGISTRY} -u "${REGISTRY_USERNAME}" -p "${REGISTRY_PASSWORD}"
fi

if [ -n "$RANCHER_AGENT_IMAGE" ]; then
  sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}
  sudo docker pull ${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}
  sudo systemctl restart rke2-server
fi

sudo mkdir -p /home/${USER}/.kube
sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config
sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube