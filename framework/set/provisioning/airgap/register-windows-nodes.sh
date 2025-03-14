#!/bin/bash

PEM_FILE=$1
WINS_PEM_FILE=$2
USER=$3
GROUP=$4
WINS_USER=$5
BASTION_IP=$6
NODE_PRIVATE_IP=$7
REGISTRATION_COMMAND=$8
REGISTRY=$9

set -e

echo ${PEM_FILE} | base64 -d | sudo tee /home/${USER}/airgap.pem > /dev/null
echo ${WINS_PEM_FILE} | base64 -d | sudo tee /home/${USER}/wins.pem > /dev/null

echo "powershell.exe ${REGISTRATION_COMMAND}" > /home/${USER}/wins_registration_command.txt

REGISTRATION_COMMAND=$(cat /home/$USER/wins_registration_command.txt)

PEM=/home/${USER}/airgap.pem
WINS_PEM=/home/${USER}/wins.pem

sudo chmod 600 ${PEM}
sudo chmod 600 /home/${USER}/wins.pem
sudo chown ${USER}:${GROUP} ${PEM} 
sudo chown ${USER}:${GROUP} /home/${USER}/wins.pem

DOCKER_DAEMON="{ \"\\\"insecure-registries\\\"\" : [ \"\\\"${REGISTRY}\\\"\" ] }"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $WINS_PEM $WINS_USER@$NODE_PRIVATE_IP \
    powershell.exe -Command New-Item -Path C:\\ProgramData\\docker\\config -ItemType Directory -Force

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $WINS_PEM $WINS_USER@$NODE_PRIVATE_IP \
    icacls C:\\ProgramData\\docker\\config /grant "Everyone:(F)"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $WINS_PEM $WINS_USER@$NODE_PRIVATE_IP \
    powershell.exe -Command "Set-Content -Path C:\\ProgramData\\docker\\config\\daemon.json -Value '$DOCKER_DAEMON'"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $WINS_PEM $WINS_USER@$NODE_PRIVATE_IP \
    powershell.exe -Command "Restart-Service docker"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $WINS_PEM $WINS_USER@$NODE_PRIVATE_IP "$REGISTRATION_COMMAND"