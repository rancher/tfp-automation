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

MAX_TOKEN_RETRIES=10
TOKEN_RETRY_DELAY=10
TOKEN_ATTEMPT=1

while [[ -z "${IMPORT_COMMAND}" && $TOKEN_ATTEMPT -le $MAX_TOKEN_RETRIES ]]; do
    echo "Waiting for registration token..."
    sleep $TOKEN_RETRY_DELAY
    IMPORT_COMMAND=$(cat /home/$USER/import_command.txt)
    TOKEN_ATTEMPT=$((TOKEN_ATTEMPT+1))
done

if [[ -z "${IMPORT_COMMAND}" ]]; then
    echo "ERROR: Registration token is still empty after waiting."
    exit 1
fi

MAX_RETRIES=5
RETRY_DELAY=15
ATTEMPT=1
SUCCESS=0

while [ $ATTEMPT -le $MAX_RETRIES ]; do
    eval "$IMPORT_COMMAND"
    EXIT_CODE=$?

    if [ $EXIT_CODE -eq 0 ]; then
        SUCCESS=1
        break
    else
        sleep $RETRY_DELAY
        ATTEMPT=$((ATTEMPT+1))
    fi
done

if [ $SUCCESS -ne 1 ]; then
    exit 1
fi