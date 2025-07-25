#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
HOSTNAME=$3
CERT_TYPE=$4
RANCHER_TAG_VERSION=$5
CHART_VERSION=$6
RANCHER_IMAGE=$7
BASTION=$8
RANCHER_AGENT_IMAGE=${9}
PROXY_PORT="3228"
NO_PROXY="localhost\\,127.0.0.0/8\\,10.0.0.0/8\\,172.0.0.0/8\\,192.168.0.0/16\\,.svc\\,.cluster.local\\,cattle-system.svc\\,169.254.169.254"

set -ex

echo "Adding Helm chart repo"
helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Upgrading Rancher"
if [ "$CERT_TYPE" == "self-signed" ]; then
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --set rancherImage=${RANCHER_IMAGE} \
                                                                                    --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                    --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                    --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                    --set noProxy="${NO_PROXY}" \
                                                                                    --set agentTLSMode=system-store \
                                                                                    --devel

    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --set rancherImage=${RANCHER_IMAGE} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                    --set noProxy="${NO_PROXY}" \
                                                                                    --set agentTLSMode=system-store \
                                                                                    --devel
    fi
elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set rancherImage=${RANCHER_IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                     --set noProxy="${NO_PROXY}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --devel
    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImage=${RANCHER_IMAGE} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                     --set noProxy="${NO_PROXY}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --devel
    fi
else
    echo "Unsupported CERT_TYPE: $CERT_TYPE"
    exit 1
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher