#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_ONE_PUBLIC_IP=$4
K3S_SERVER_ONE_PRIVATE_IP=$5
K3S_NEW_SERVER_PRIVATE_IP=$6
K3S_TOKEN=$7
REGISTRY_USERNAME=$8
REGISTRY_PASSWORD=$9
PEM_FILE=/home/$USER/airgap.pem

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export K3S_SERVER_ONE_PUBLIC_IP=${K3S_SERVER_ONE_PUBLIC_IP}; \
   export K3S_SERVER_ONE_PRIVATE_IP=${K3S_SERVER_ONE_PRIVATE_IP}; \
   export K3S_NEW_SERVER_PRIVATE_IP=${K3S_NEW_SERVER_PRIVATE_IP}; \
   export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
   export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; \
   export K3S_TOKEN=${K3S_TOKEN}; $cmd"
}

setupRegistry() {
  sudo mkdir -p /etc/rancher/k3s
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

registryFunction=$(declare -f setupRegistry)
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "${registryFunction}; setupRegistry"
runSSH "${K3S_NEW_SERVER_PRIVATE_IP}" "sudo INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_EXEC=\"agent --server https://[${K3S_SERVER_ONE_PUBLIC_IP}]:6443\" INSTALL_K3S_SKIP_DOWNLOAD=true sh /home/${USER}/install.sh"