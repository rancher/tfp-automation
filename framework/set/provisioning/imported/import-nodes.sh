#!/bin/bash

PEM_FILE=$1
USER=$2
GROUP=$3
NODE_ONE_PUBLIC_DNS=$4
IMPORT_COMMAND=$5
RKE_KUBE_CONFIG_FILE=${6}

set -ex

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl"
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

eval "$IMPORT_COMMAND"