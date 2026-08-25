#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
HOSTNAME=$3
RANCHER_TAG_VERSION=$4
CHART_VERSION=$5
RANCHER_IMAGE=$6
RANCHER_AGENT_IMAGE=${7}
UPGRADED_MCM=${8}

if [[ $RANCHER_TAG_VERSION == v2.11* || $RANCHER_TAG_VERSION == v2.10* ]]; then
    RANCHER_TAG="--set rancherImageTag=${RANCHER_TAG_VERSION}" 
    IMAGE="--set rancherImage=${RANCHER_IMAGE}"
    VERSION="--version ${CHART_VERSION}"
else
    IMAGE_REGISTRY="${RANCHER_IMAGE%%/*}"

    if [[ -n "$RANCHER_AGENT_IMAGE" || "$RANCHER_IMAGE" == registry* ]]; then
        IMAGE_REPOSITORY="rancher/rancher"
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
    if [ "$REPO" == "prime-release" ]; then
        helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}
    else
        helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
    fi
}

upgrade_mcm_off() {
    echo "Upgrading Rancher with MCM off"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[1].value=prime' \
                                                                                        --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[2].value=suse' \
                                                                                        --set 'extraEnv[3].name=CATTLE_FEATURES' \
                                                                                        --set 'extraEnv[3].value=multi-cluster-management=false' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set ingress.tls.source=secret \
                                                                                        --devel

    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                        --set 'extraEnv[0].value=multi-cluster-management=false' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set ingress.tls.source=secret \
                                                                                        --devel
    fi
}

upgrade_prime_head_rancher() {
    echo "Upgrading Rancher"
    helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${VERSION} \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
}

upgrade_default_rancher() {
    echo "Upgrading Rancher"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[1].value=prime' \
                                                                                        --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[2].value=suse' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set ingress.tls.source=secret \
                                                                                        --devel

    else
        helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
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

if [[ $REPO == "prime-release" ]]; then
    . /etc/os-release

    [[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt -y install yq
    [[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install yq -y
    [[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install  -y yq

    HEAD_SERIES=""
    if [[ $RANCHER_TAG_VERSION =~ ^v([0-9]+\.[0-9]+)-head$ ]]; then
        HEAD_SERIES="${BASH_REMATCH[1]}"
    fi

    if [[ -z "$HEAD_SERIES" && $RANCHER_TAG_VERSION == "head" ]]; then
      HEAD_SERIES="${RANCHER_HEAD_SERIES:-}"
      if [[ -z "$HEAD_SERIES" ]]; then
        HEAD_SERIES=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher[].version' | sed -nE 's/^([0-9]+\.[0-9]+)\.[0-9]+.*-head$/\1/p' | sort -Vu | tail -n 1)
      fi
    fi

    if [[ $RANCHER_TAG_VERSION == "head" && -z "$HEAD_SERIES" ]]; then
        LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher | map(select(.version | test("^"))) | sort_by(.created) | .[-1].version')
        VERSION="--version ${LATEST_CHART_VERSION}"
    fi

    if [[ -n "$HEAD_SERIES" ]]; then
        SERIES_FILTER="^${HEAD_SERIES//./\\\\.}"
        LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r ".entries.rancher | map(select(.version | test(\"${SERIES_FILTER}\"))) | sort_by(.created) | .[-1].version")
        VERSION="--version ${LATEST_CHART_VERSION}"
    fi
fi

# Needed to get the latest chart version if RANCHER_TAG_VERSION contains "head"
if [[ $RANCHER_TAG_VERSION == *head* && $REPO != "prime-release" ]]; then
    LATEST_CHART_VERSION=$(helm search repo upgraded-rancher-${REPO} --devel | tail -n +2 | head -n 1 | cut -f2)
    VERSION="--version ${LATEST_CHART_VERSION}"
fi

if [ -n "$UPGRADED_MCM" ]; then
    case "$UPGRADED_MCM" in
        "false"|"toggledOn")
            upgrade_mcm_off
            ;;
        *)
            upgrade_default_rancher
            ;;
    esac
elif [[ $REPO == "prime-release" ]]; then
    upgrade_prime_head_rancher
else
    upgrade_default_rancher
fi

wait_for_rollout
wait_for_rancher