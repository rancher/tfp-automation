#!/bin/bash

WINS_PEM_FILE=$1
USER=$2
GROUP=$3
WINS_USER=$4
NODE_PRIVATE_IP=$5
REGISTRATION_COMMAND=$6
REGISTRY=$7

echo ${WINS_PEM_FILE} | base64 -d | sudo tee /home/${USER}/wins.pem > /dev/null
echo "powershell.exe ${REGISTRATION_COMMAND}" > /home/${USER}/wins_registration_command.txt
echo "{ \"insecure-registries\" : [ \"${REGISTRY}\" ] }" | sudo tee /tmp/daemon.json

REGISTRATION_COMMAND=$(cat /home/$USER/wins_registration_command.txt)
WINS_PEM=/home/${USER}/wins.pem

sudo chmod 600 ${WINS_PEM}
sudo chown ${USER}:${GROUP} ${WINS_PEM}

runSSH() {
  local server="$1"
  local cmd="$2"
  
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=300 -o ServerAliveInterval=30 -o ServerAliveCountMax=10 -o TCPKeepAlive=yes -i "$WINS_PEM" "$WINS_USER@$server" \
  "$cmd"
}

runSSH "${NODE_PRIVATE_IP}" "powershell.exe -Command New-Item -Path C:\\ProgramData\\docker\\config -ItemType Directory -Force"
runSSH "${NODE_PRIVATE_IP}" "icacls C:\\ProgramData\\docker\\config /grant \"Everyone:(F)\""
scp -i "${WINS_PEM}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null /tmp/daemon.json "${WINS_USER}@${NODE_PRIVATE_IP}:C:/ProgramData/docker/config/daemon.json"
runSSH "${NODE_PRIVATE_IP}" "${REGISTRATION_COMMAND}" || true

# Fail-safe sleep to ensure the Windows registration command has time to fully finish
sleep 70