#!/bin/bash

USER=$1
GROUP=$2
RKE2_SERVER_ONE_IP=$3
RKE2_NEW_SERVER_IP=$4
RKE2_NEW_SERVER_PRIVATE_IP=$5
HOSTNAME=$6
CNI=$7
REGISTRY_USERNAME=$8
REGISTRY_PASSWORD=$9
RKE2_TOKEN=${10}
RANCHER_IMAGE=${11}
RANCHER_TAG_VERSION=${12}
CLUSTER_CIDR=${13}
SERVICE_CIDR=${14}
PEM_FILE=/home/$USER/airgap.pem

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export RKE2_SERVER_ONE_IP=${RKE2_SERVER_ONE_IP}; \
   export RKE2_NEW_SERVER_IP=${RKE2_NEW_SERVER_IP}; \
   export HOSTNAME=${HOSTNAME}; \
   export CNI=${CNI}; \
   export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
   export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; \
   export RKE2_TOKEN=${RKE2_TOKEN}; \
   export CLUSTER_CIDR=${CLUSTER_CIDR}; \
   export SERVICE_CIDR=${SERVICE_CIDR}; $cmd"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/rke2
    sudo tee /etc/rancher/rke2/config.yaml > /dev/null << EOF
server: https://[$RKE2_SERVER_ONE_IP]:9345
write-kubeconfig-mode: 644
node-ip: ${RKE2_NEW_SERVER_IP}
node-external-ip: ${RKE2_NEW_SERVER_IP}
cni: ${CNI}
token: ${RKE2_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
tls-san:
  - ${HOSTNAME}
EOF
}

setupRegistry() {
  sudo tee /etc/rancher/rke2/registries.yaml > /dev/null << EOF
mirrors:
  docker.io:
    endpoint:
      - "https://registry-1.docker.io"
configs:
  "registry-1.docker.io":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
  "docker.io":
    auth:
      username: "${REGISTRY_USERNAME}"
      password: "${REGISTRY_PASSWORD}"
EOF
}

configFunction=$(declare -f setupConfig)
runSSH "${RKE2_NEW_SERVER_PRIVATE_IP}" "${configFunction}; setupConfig"

registryFunction=$(declare -f setupRegistry)
runSSH "${RKE2_NEW_SERVER_PRIVATE_IP}" "${registryFunction}; setupRegistry"

runSSH "${RKE2_NEW_SERVER_PRIVATE_IP}" "sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh"
runSSH "${RKE2_NEW_SERVER_PRIVATE_IP}" "sudo systemctl enable rke2-server"
runSSH "${RKE2_NEW_SERVER_PRIVATE_IP}" "sudo systemctl start rke2-server"