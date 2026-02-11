#!/bin/bash

set -euo pipefail

export PREFIX="${PREFIX}"
SUBSCRIPTION_ID=$(az account show --query id -o tsv)

echo "Cleanup in progress..."

RESOURCE_GROUPS=$(az group list --query "[?starts_with(name, '${PREFIX}')].name" -o tsv)

if [ -n "$RESOURCE_GROUPS" ]; then
  while read -r rg; do
    echo "Deleting resource group: $rg"
    az group delete \
      --name "$rg" \
      --subscription "$SUBSCRIPTION_ID" \
      --yes \
      --no-wait
  done <<< "$RESOURCE_GROUPS"
else
  echo "No matching resource groups found."
fi

AKS_RESOURCE_GROUPS=$(az group list --query "[?starts_with(name, 'MC_${PREFIX}')].name" -o tsv)

if [ -n "$AKS_RESOURCE_GROUPS" ]; then
  while read -r rg; do
    echo "Deleting AKS resource group: $rg"
    az group delete \
      --name "$rg" \
      --subscription "$SUBSCRIPTION_ID" \
      --yes \
      --no-wait
  done <<< "$AKS_RESOURCE_GROUPS"
else
  echo "No matching AKS resource groups found."
fi

echo "Cleanup completed!"