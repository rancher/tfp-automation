#!/bin/bash

USER=$1
GROUP=$2
VPC_IP=$3
RKE2_SERVER_ONE_IP=$4
RKE2_TOKEN=$5
REGISTRY=$6
REGISTRY_USERNAME=$7
REGISTRY_PASSWORD=$8
RANCHER_IMAGE=$9
RANCHER_TAG_VERSION=${10}
RANCHER_AGENT_IMAGE=${11}
PEM_FILE=/home/$USER/airgap.pem

set -e

run_ssh() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM_FILE" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export VPC_IP=${VPC_IP}; \
   export RKE2_SERVER_ONE_IP=${RKE2_SERVER_ONE_IP}; \
   export RKE2_TOKEN=${RKE2_TOKEN}; \
   export REGISTRY=${REGISTRY}; \
   export REGISTRY_USERNAME=${REGISTRY_USERNAME}; \
   export REGISTRY_PASSWORD=${REGISTRY_PASSWORD}; $cmd"
}

setup_config() {
    sudo mkdir -p /etc/rancher/rke2
    sudo tee /etc/rancher/rke2/config.yaml > /dev/null << EOF
token: ${RKE2_TOKEN}
system-default-registry: ${REGISTRY}
tls-san:
  - ${RKE2_SERVER_ONE_IP}
EOF
}

setup_registry() {
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
}

setup_docker_daemon() {
  sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
  "insecure-registries" : [ "${REGISTRY}" ]
}
EOF
}

setup_networking() {
  sudo systemctl disable systemd-resolved; sudo systemctl stop systemd-resolved
  sudo sed -i.bak "s/^nameserver .*/nameserver ${VPC_IP}/" /etc/resolv.conf
  sudo sed -i.bak "/^options /d" /etc/resolv.conf
  echo "options edns0" | sudo tee -a /etc/resolv.conf

  sudo tee /etc/NetworkManager/conf.d/90-dns-none.conf > /dev/null << EOF
[main]
dns=none
EOF
}

run_ssh "${RKE2_SERVER_ONE_IP}" "sudo mv /home/${USER}/kubectl /usr/local/bin/"

configFunction=$(declare -f setup_config)
run_ssh "${RKE2_SERVER_ONE_IP}" "${configFunction}; setup_config"

setupRegistryFunction=$(declare -f setup_registry)
run_ssh "${RKE2_SERVER_ONE_IP}" "${setupRegistryFunction}; setup_registry"

run_ssh "${RKE2_SERVER_ONE_IP}" "sudo INSTALL_RKE2_ARTIFACT_PATH=/home/${USER} sh install.sh"
run_ssh "${RKE2_SERVER_ONE_IP}" "sudo systemctl enable rke2-server"
run_ssh "${RKE2_SERVER_ONE_IP}" "sudo systemctl start rke2-server"

setupDaemonFunction=$(declare -f setup_docker_daemon)
run_ssh "${RKE2_SERVER_ONE_IP}" "${setupDaemonFunction}; setup_docker_daemon"
run_ssh "${RKE2_SERVER_ONE_IP}" "sudo systemctl restart docker && sudo systemctl daemon-reload"

setupNetworkingFunction=$(declare -f setup_networking)
run_ssh "${RKE2_SERVER_ONE_IP}" "${setupNetworkingFunction}; setup_networking"

if [ -n "$RANCHER_AGENT_IMAGE" ]; then
  run_ssh "${RKE2_SERVER_ONE_IP}" "sudo docker pull ${REGISTRY}/${RANCHER_IMAGE}:${RANCHER_TAG_VERSION}"
  run_ssh "${RKE2_SERVER_ONE_IP}" "sudo docker pull ${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}"
  run_ssh "${RKE2_SERVER_ONE_IP}" "sudo systemctl restart rke2-server"
fi

run_ssh "${RKE2_SERVER_ONE_IP}" "sudo mkdir -p /home/${USER}/.kube"
run_ssh "${RKE2_SERVER_ONE_IP}" "sudo cp /etc/rancher/rke2/rke2.yaml /home/${USER}/.kube/config"
run_ssh "${RKE2_SERVER_ONE_IP}" "sudo chown -R ${USER}:${GROUP} /home/${USER}/.kube"

mkdir -p ~/.kube
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${PEM_FILE} ${USER}@${RKE2_SERVER_ONE_IP} "sudo cat /home/${USER}/.kube/config" > ~/.kube/config
sed -i "s|server: https://127.0.0.1:6443|server: https://${RKE2_SERVER_ONE_IP}:6443|" ~/.kube/config
kubectl get nodes