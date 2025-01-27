#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
HOSTNAME=$3
RANCHER_TAG_VERSION=$4
RANCHER_IMAGE=$5
BASTION=$6
STAGING_RANCHER_AGENT_IMAGE=${7}
PROXY_PORT="3128"
NO_PROXY="localhost\\,127.0.0.0/8\\,10.0.0.0/8\\,172.0.0.0/8\\,192.168.0.0/16\\,.svc\\,.cluster.local\\,cattle-system.svc\\,169.254.169.254"

set -ex

echo "Adding Helm chart repo"
helm repo add rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Upgrading Rancher"
if [ -n "$STAGING_RANCHER_AGENT_IMAGE" ]; then
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set rancherImage=${RANCHER_IMAGE} \
                                                                                 --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                 --set "extraEnv[0].value=${STAGING_RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                 --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                 --set noProxy="${NO_PROXY}" --devel

else
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImage=${RANCHER_IMAGE} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                 --set noProxy="${NO_PROXY}"
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher