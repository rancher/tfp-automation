#!/bin/bash

PEM_FILE=$1
USER=$2
GROUP=$3
NODE_ONE_PUBLIC_DNS=$4
IMPORT_COMMAND=$5
RKE_KUBE_CONFIG_FILE=${6}

set -ex

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
mkdir -p ~/.kube
rm kubectl

if [ -n "${RKE_KUBE_CONFIG_FILE}" ]; then
    echo "${RKE_KUBE_CONFIG_FILE}" > /home/$USER/.kube/config
fi

echo ${PEM_FILE} | sudo base64 -d > /home/${USER}/key.pem
echo "${IMPORT_COMMAND}" > /home/${USER}/import_command.txt
IMPORT_COMMAND=$(cat /home/$USER/import_command.txt)

PEM=/home/${USER}/key.pem
sudo chmod 600 ${PEM}
sudo chown ${USER}:${GROUP} ${PEM}

ssh -o ProxyCommand="ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM -W %h:%p $USER@$NODE_ONE_PUBLIC_DNS" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i $PEM $USER@$NODE_ONE_PUBLIC_DNS "$IMPORT_COMMAND"