#!/bin/bash

USER=$1
GROUP=$2
REGISTRY_USERNAME=$3
REGISTRY_PASSWORD=$4
K8S_VERSION=$5
RKE2_SERVER_ONE_IP=$6
RKE2_SERVER_TWO_IP=$7
RKE2_SERVER_THREE_IP=$8
DOCKER_DIR="/etc/systemd/system/docker.service.d"
PORT="3228"

set -e

echo "Logging into the private registry..."
sudo docker login https://registry-1.docker.io -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD}

echo "Starting proxy..."
sudo mkdir -p /home/$USER/squid
PROXY_DIR=/home/$USER/squid
sudo mv /tmp/squid.conf ${PROXY_DIR}/squid.conf

sudo mkdir -p /var/cache/squid
sudo chown -R ${USER}:${GROUP} /var/cache/squid
sudo chmod 777 /var/cache/squid

sudo docker run -d -v ${PROXY_DIR}/squid.conf:/etc/squid/squid.conf -v /var/cache/squid:/var/cache/squid -p ${PORT}:${PORT} ubuntu/squid

sudo mv /tmp/keyfile.pem /home/$USER/keyfile.pem
PEM=/home/$USER/keyfile.pem
sudo chown $USER:$GROUP $PEM
chmod 600 $PEM

curl -fsSL --max-time 30 -o rke2.linux-amd64.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-amd64.tar.gz
curl -fsSL --max-time 30 -o rke2.linux-arm64.tar.gz https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2.linux-arm64.tar.gz
curl -fsSL --max-time 30 -o rke2-images.linux-amd64.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-amd64.tar.zst
curl -fsSL --max-time 30 -o rke2-images.linux-arm64.tar.zst https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/rke2-images.linux-arm64.tar.zst
curl -fsSL --max-time 30 -o sha256sum-amd64.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-amd64.txt
curl -fsSL --max-time 30 -o sha256sum-arm64.txt https://github.com/rancher/rke2/releases/download/${K8S_VERSION}+rke2r1/sha256sum-arm64.txt

curl -fsSL --max-time 30 -o install.sh https://get.rke2.io
chmod +x install.sh

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Validating checksum for rke2-images.linux-${ARCH}.tar.zst"
ZIP_NAME="rke2-images.linux-${ARCH}.tar.zst"
CHECKSUM_LINE=$(grep "${ZIP_NAME}" /home/${USER}/sha256sum-${ARCH}.txt)

if [ -z "$CHECKSUM_LINE" ]; then
  echo "ERROR: Checksum for $ZIP_NAME not found in sha256sum-${ARCH}.txt file!"
  exit 1
fi

CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
echo "$CHECKSUM  /home/${USER}/rke2-images.linux-${ARCH}.tar.zst" | sha256sum -c -

echo "Installing kubectl"
KUBECTL_VERSION="v1.36.0"
curl -fsSL --max-time 30 -o kubectl https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl
curl -fsSL --max-time 30 -o kubectl.sha256 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256
echo "$(cat kubectl.sha256) kubectl" | sha256sum -c
sudo chmod +x kubectl

echo "Copying files to RKE2 server one"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null kubectl ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/

echo "Copying files to RKE2 server two"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/

echo "Copying files to RKE2 server three"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/

sudo mv kubectl /usr/local/bin/