#!/bin/bash

RESOURCE_PREFIX=$1
ZONE=$2
MACHINE_TYPE=$3
PROJECT_ID=$4
SERVICE_ACCOUNT_KEY_FILE=$5

set -e

base64 -d <<< $SERVICE_ACCOUNT_KEY_FILE > /home/$USER/key.json
KEY=/home/$USER/key.json
chmod 600 $KEY

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" || $ARCH == "amd64" ]]; then
    ARCH="x86_64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm"
fi

echo "Installing gcloud..."
curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-${ARCH}.tar.gz
tar -xf google-cloud-cli-linux-${ARCH}.tar.gz
./google-cloud-sdk/install.sh --quiet > /dev/null
rm google-cloud-cli-linux-${ARCH}.tar.gz

export PATH="$PWD/google-cloud-sdk/bin:$PATH"

echo "Logging into project..."
gcloud config set project ${PROJECT_ID} > /dev/null 2>&1

echo "Authenticating with service account..."
gcloud auth activate-service-account --key-file=${KEY} > /dev/null 2>&1

echo "Installing GKE plugin..."
gcloud components install gke-gcloud-auth-plugin --quiet > /dev/null 2>&1

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

echo "Creating GKE cluster..."
gcloud container clusters create ${RESOURCE_PREFIX} --zone=${ZONE} --num-nodes=3 \
                                                                       --cluster-version=1.33 \
                                                                       --machine-type=${MACHINE_TYPE} \
                                                                       --disk-size=100 \
                                                                       --release-channel=regular > /dev/null 2>&1