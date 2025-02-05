#!/bin/bash

PEM_FILE=$1
USER=$2
GROUP=$3
BASTION_IP=$4
NODE_PRIVATE_IP=$5
REGISTRATION_COMMAND=$6
REGISTRY=$7

set -e

echo ${PEM_FILE} | sudo base64 -d > /home/${USER}/airgap.pem
echo "${REGISTRATION_COMMAND}" > /home/${USER}/registration_command.txt
REGISTRATION_COMMAND=$(cat /home/$USER/registration_command.txt)

PEM=/home/${USER}/airgap.pem
sudo chmod 600 ${PEM}
sudo chown ${USER}:${GROUP} ${PEM}

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM $USER@$NODE_PRIVATE_IP "sudo tee /etc/docker/daemon.json > /dev/null << EOF
{
    \"insecure-registries\" : [ \"${REGISTRY}\" ]
    }
EOF"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM $USER@$NODE_PRIVATE_IP \
    "sudo systemctl restart docker && sudo systemctl daemon-reload"

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$BASTION_IP" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM $USER@$NODE_PRIVATE_IP "$REGISTRATION_COMMAND"