#!/bin/bash

USER=$1
GROUP=$2
RKE2_SERVER_ONE_IP=$3
RKE2_NEW_SERVER_IP=$4
PUBLIC_FQDN=$5
INTERNAL_FQDN=$6
RKE2_TOKEN=$7
RANCHER_IMAGE=$8
RANCHER_TAG_VERSION=$9
CLUSTER_CIDR=${10}
SERVICE_CIDR=${11}
PEM_FILE=/home/$USER/airgap.pem

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export RKE2_SERVER_ONE_IP=${RKE2_SERVER_ONE_IP}; \
   export PUBLIC_FQDN=${PUBLIC_FQDN}; \
   export INTERNAL_FQDN=${INTERNAL_FQDN}; \
   export RKE2_TOKEN=${RKE2_TOKEN}; \
   export REGISTRY=${REGISTRY}; \
   export CLUSTER_CIDR=${CLUSTER_CIDR}; \
   export SERVICE_CIDR=${SERVICE_CIDR}; $cmd"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/rke2
    sudo tee /etc/rancher/rke2/config.yaml > /dev/null << EOF
cni: calico
server: https://${RKE2_SERVER_ONE_IP}:9345
token: ${RKE2_TOKEN}
tls-san:
  - ${PUBLIC_FQDN}
  - ${INTERNAL_FQDN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
EOF
}

configFunction=$(declare -f setupConfig)
runSSH "${RKE2_NEW_SERVER_IP}" "${configFunction}; setupConfig"

runSSH "${RKE2_NEW_SERVER_IP}" "sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh"
runSSH "${RKE2_NEW_SERVER_IP}" "sudo systemctl enable rke2-server"
runSSH "${RKE2_NEW_SERVER_IP}" "sudo systemctl start rke2-server"

kubectl get nodes