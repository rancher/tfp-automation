#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
HOSTNAME=$3
RANCHER_TAG_VERSION=$4
CHART_VERSION=$5
RANCHER_IMAGE=$6
BASTION=$7
RANCHER_AGENT_IMAGE=${8}
PROXY_PORT="3228"
NO_PROXY="localhost\\,127.0.0.0/8\\,10.0.0.0/8\\,172.0.0.0/8\\,192.168.0.0/16\\,.svc\\,.cluster.local\\,cattle-system.svc\\,169.254.169.254"

if [[ $RANCHER_TAG_VERSION == v2.11* || $RANCHER_TAG_VERSION == v2.10* ]]; then
    RANCHER_TAG="--set rancherImageTag=${RANCHER_TAG_VERSION}" 
    IMAGE="--set rancherImage=${RANCHER_IMAGE}"
    VERSION="--version ${CHART_VERSION}"
else
    IMAGE_REGISTRY="${RANCHER_IMAGE%%/*}"

    if [[ -n "$RANCHER_AGENT_IMAGE" || "$RANCHER_IMAGE" == registry* ]]; then
        IMAGE_REPOSITORY="rancher"
    else
        IMAGE_REPOSITORY="${RANCHER_IMAGE#*/}"
    fi
    
    RANCHER_TAG="--set image.tag=${RANCHER_TAG_VERSION}"
    IMAGE="--set image.repository=${IMAGE_REPOSITORY} --set image.registry=${IMAGE_REGISTRY}"
    VERSION="--version ${CHART_VERSION}"
fi

set -ex

setup_helm_repo() {
    echo "Adding Helm chart repo"
    helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
}

upgrade_rancher() {
    echo "Upgrading Rancher"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set 'extraEnv[0].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[0].value=prime' \
                                                                                        --set 'extraEnv[1].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[1].value=suse' \
                                                                                        --set 'extraEnv[2].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[2].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                        --set noProxy="${NO_PROXY}" \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set ingress.tls.source=secret \
                                                                                        --devel

    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                        --set noProxy="${NO_PROXY}" \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set ingress.tls.source=secret \
                                                                                        --devel
    fi
}

wait_for_rollout() {
    echo "Waiting for Rancher to be rolled out"
    kubectl -n cattle-system rollout status deploy/rancher
    kubectl -n cattle-system get deploy rancher
}

wait_for_rancher() {
    echo "Waiting 15 seconds to be able to login to Rancher"
    sleep 15
}

setup_helm_repo

# Needed to get the latest chart version if RANCHER_TAG_VERSION contains "head"
if [[ $RANCHER_TAG_VERSION == *head* ]]; then
    LATEST_CHART_VERSION=$(helm search repo upgraded-rancher-${REPO} --devel | tail -n +2 | head -n 1 | cut -f2)
    VERSION="--version ${LATEST_CHART_VERSION}"
fi

upgrade_rancher
wait_for_rollout
wait_for_rancher