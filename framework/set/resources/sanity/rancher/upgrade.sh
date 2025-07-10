#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_TYPE=$3
HOSTNAME=$4
RANCHER_TAG_VERSION=$5
RANCHER_IMAGE=$6
RANCHER_AGENT_IMAGE=${7}

set -ex

echo "Adding Helm chart repo"
helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Upgrading Rancher"
if [ "$CERT_TYPE" == "self-signed" ]; then
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --set rancherImage=${RANCHER_IMAGE} \
                                                                                    --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                    --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                    --devel

    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --set rancherImage=${RANCHER_IMAGE} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --devel
    fi
elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
    RAND_STR=$(LC_ALL=C tr -dc 'a-z0-9' </dev/urandom | head -c 12)
    LETS_ENCRYPT_EMAIL="${RAND_STR}@gmail.com"

    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set rancherImage=${RANCHER_IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImage=${RANCHER_IMAGE} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --devel
    fi
else
    echo "Unsupported CERT_TYPE: $CERT_TYPE"
    exit 1
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher