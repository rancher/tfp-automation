#!/bin/bash

RESOURCE_PREFIX=$1
LOCATION=$2
NODE_COUNT=$3
NODE_SIZE=$4
AZURE_CLIENT_ID=$5
AZURE_CLIENT_SECRET=$6
AZURE_TENANT_ID=$7
USER=$8
PUB_FILE=${9}

RESOURCE_GROUP_NAME="${RESOURCE_PREFIX}-rg"

set -e

base64 -d <<< $PUB_FILE > /home/$USER/cloud.pub
PUB=/home/$USER/cloud.pub
chmod 600 $PUB

. /etc/os-release

[[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && \
    sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release && \
    sudo mkdir -p /etc/apt/keyrings && curl -sLS https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | sudo tee /etc/apt/keyrings/microsoft.gpg > /dev/null && \
    sudo chmod go+r /etc/apt/keyrings/microsoft.gpg && \
    AZ_DIST=$(lsb_release -cs)
echo "Types: deb
URIs: https://packages.microsoft.com/repos/azure-cli/
Suites: ${AZ_DIST}
Components: main
Architectures: $(dpkg --print-architecture)
Signed-by: /etc/apt/keyrings/microsoft.gpg" | sudo tee /etc/apt/sources.list.d/azure-cli.sources > /dev/null && \
    sudo apt-get update && sudo apt-get install azure-cli -y > /dev/null 2>&1
[[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install -y azure-cli > /dev/null 2>&1

az login --service-principal -u "$AZURE_CLIENT_ID" -p "$AZURE_CLIENT_SECRET" --tenant "$AZURE_TENANT_ID" > /dev/null 2>&1

echo "Creating AKS cluster..."
az aks create --resource-group "$RESOURCE_GROUP_NAME" --name "$RESOURCE_PREFIX" --node-count "$NODE_COUNT" --node-vm-size "$NODE_SIZE" \
              --enable-managed-identity --ssh-key-value "$PUB" > /dev/null 2>&1

echo "Getting AKS cluster credentials..."
az aks get-credentials --resource-group "$RESOURCE_GROUP_NAME" --name "$RESOURCE_PREFIX" > /dev/null 2>&1