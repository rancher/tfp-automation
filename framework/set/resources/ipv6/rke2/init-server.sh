#!/bin/bash

USER=$1
GROUP=$2
RKE2_SERVER_ONE_IP=$3
PUBLIC_FQDN=$4
INTERNAL_FQDN=$5
RKE2_TOKEN=$6
RANCHER_IMAGE=$7
RANCHER_TAG_VERSION=$8
CLUSTER_CIDR=${9}
SERVICE_CIDR=${10}
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
   export CLUSTER_CIDR=${CLUSTER_CIDR}; \
   export SERVICE_CIDR=${SERVICE_CIDR}; $cmd"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/rke2
    sudo tee /etc/rancher/rke2/config.yaml > /dev/null << EOF
cni: calico
token: ${RKE2_TOKEN}
tls-san:
  - ${PUBLIC_FQDN}
  - ${INTERNAL_FQDN}
cluster-cidr: ${CLUSTER_CIDR}
service-cidr: ${SERVICE_CIDR}
EOF
}

runSSH "${RKE2_SERVER_ONE_IP}" "sudo mv /home/${USER}/kubectl /usr/local/bin/"

configFunction=$(declare -f setupConfig)
runSSH "${RKE2_SERVER_ONE_IP}" "${configFunction}; setupConfig"

runSSH "${RKE2_SERVER_ONE_IP}" "sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh"
runSSH "${RKE2_SERVER_ONE_IP}" "sudo systemctl enable rke2-server"
runSSH "${RKE2_SERVER_ONE_IP}" "sudo systemctl start rke2-server"

runSSH "${RKE2_SERVER_ONE_IP}" "sudo mkdir -p /home/${USER}/.kube"
runSSH "${RKE2_SERVER_ONE_IP}" "sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config"
runSSH "${RKE2_SERVER_ONE_IP}" "sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube"

mkdir -p ~/.kube
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${PEM_FILE} ${USER}@${RKE2_SERVER_ONE_IP} "sudo cat /home/${USER}/.kube/config" > ~/.kube/config
sed -i "s|server: https://127.0.0.1:6443|server: https://${RKE2_SERVER_ONE_IP}:6443|" ~/.kube/config
kubectl get nodes