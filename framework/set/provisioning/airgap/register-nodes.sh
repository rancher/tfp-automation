#!/bin/bash

PEM_FILE=$1
USER=$2
GROUP=$3
NODE_PRIVATE_IP=$4
REGISTRATION_COMMAND=$5
REGISTRY=$6

set -e

echo ${PEM_FILE} | sudo base64 -d > /home/${USER}/airgap.pem
echo "${REGISTRATION_COMMAND}" > /home/${USER}/registration_command.txt
REGISTRATION_COMMAND=$(cat /home/$USER/registration_command.txt)

PEM=/home/${USER}/airgap.pem
sudo chmod 600 ${PEM}
sudo chown ${USER}:${GROUP} ${PEM}

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$PEM" "$USER@$server" \
  "export USER=${USER}; \
   export GROUP=${GROUP}; \
   export NODE_PRIVATE_IP=${NODE_PRIVATE_IP}; \
   export REGISTRY=${REGISTRY}; \
   export REGISTRATION_COMMAND=${REGISTRATION_COMMAND}; \
   export REGISTRY=${REGISTRY}; $cmd"
}

setupDockerDaemon() {
  echo "{ \"insecure-registries\" : [ \"${REGISTRY}\" ] }" | sudo tee /etc/docker/daemon.json > /dev/null
}

dockerDaemonFunction=$(declare -f setupDockerDaemon)
runSSH "${NODE_PRIVATE_IP}" "${dockerDaemonFunction}; setupDockerDaemon"

runSSH "${NODE_PRIVATE_IP}" "sudo systemctl daemon-reload && sudo systemctl restart docker"
runSSH "${NODE_PRIVATE_IP}" "${REGISTRATION_COMMAND}"