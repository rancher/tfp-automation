#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
RKE2_SERVER_ONE_IP=$4
RKE2_NEW_SERVER_IP=$5
RKE2_TOKEN=$6
BASTION=$7
PORT="3228"
PEM_FILE=/home/$USER/keyfile.pem

set -e

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export RKE2_SERVER_ONE_IP=${RKE2_SERVER_ONE_IP}; \
   export RKE2_TOKEN=${RKE2_TOKEN}; \
   export BASTION=${BASTION}; \
   export PORT=${PORT}; ${cmd}"
}

setupConfig() {
    sudo mkdir -p /etc/rancher/rke2
    sudo tee /etc/rancher/rke2/config.yaml > /dev/null << EOF
server: https://${RKE2_SERVER_ONE_IP}:9345
token: ${RKE2_TOKEN}
tls-san:
  - ${RKE2_SERVER_ONE_IP}
EOF
}

setupProxy() {
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
}

configFunction=$(declare -f setupConfig)
runSSH "${RKE2_NEW_SERVER_IP}" "${configFunction}; setupConfig"

setupProxyFunction=$(declare -f setupProxy)
runSSH "${RKE2_NEW_SERVER_IP}" "${setupProxyFunction}; setupProxy"

runSSH "${RKE2_NEW_SERVER_IP}" "sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh"
runSSH "${RKE2_NEW_SERVER_IP}" "sudo systemctl enable rke2-server"
runSSH "${RKE2_NEW_SERVER_IP}" "sudo systemctl start rke2-server"