#!/bin/bash

USER=$1
GROUP=$2
K3S_SERVER_ONE_IP=$3
K3S_NEW_SERVER_PRIVATE_IP=$4
HOSTNAME=$5
REGISTRY_USERNAME=$6
REGISTRY_PASSWORD=$7
K3S_TOKEN=${8}
RANCHER_IMAGE=${9}
RANCHER_TAG_VERSION=${10}
CLUSTER_CIDR=${11}
SERVICE_CIDR=${12}
PEM_FILE=/home/$USER/airgap.pem

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export K3S_SERVER_ONE_IP=${K3S_SERVER_ONE_IP}; \
   export K3S_NEW_SERVER_PRIVATE_IP=${K3S_NEW_SERVER_PRIVATE_IP}; \
   export HOSTNAME=${HOSTNAME}; \
   export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
   export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; \
   export K3S_TOKEN=${K3S_TOKEN}; \
   export CLUSTER_CIDR=${CLUSTER_CIDR}; \
   export SERVICE_CIDR=${SERVICE_CIDR}; $cmd"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/k3s
    sudo tee /etc/rancher/k3s/config.yaml > /dev/null << EOF
server: https://[$K3S_SERVER_ONE_IP]:6443
write-kubeconfig-mode: 644
token: ${K3S_TOKEN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
tls-san:
  - ${HOSTNAME}
EOF
}

setupRegistry() {
  sudo tee /etc/rancher/k3s/registries.yaml > /dev/null << EOF
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

runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo mv /home/${USER}/k3s /usr/local/bin/"

configFunction=$(declare -f setupConfig)
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "${configFunction}; setupConfig"

registryFunction=$(declare -f setupRegistry)
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "${registryFunction}; setupRegistry"

runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo INSTALL_K3S_SKIP_DOWNLOAD=true sh /home/${USER}/install.sh"
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo systemctl enable k3s"
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo systemctl start k3s"