#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_MANAGER_VERSION=$3
CERT_TYPE=$4
HOSTNAME=$5
RANCHER_TAG_VERSION=$6
CHART_VERSION=$7
BOOTSTRAP_PASSWORD=$8
RANCHER_IMAGE=$9
BASTION=${10}
RANCHER_AGENT_IMAGE=${11}
PROXY_PORT="3228"
NO_PROXY="localhost\\,127.0.0.0/8\\,10.0.0.0/8\\,172.0.0.0/8\\,192.168.0.0/16\\,.svc\\,.cluster.local\\,cattle-system.svc\\,169.254.169.254"

if [[ $RANCHER_TAG_VERSION == v2.11* ]]; then
    RANCHER_TAG="--set rancherImageTag=${RANCHER_TAG_VERSION}" 
    IMAGE="--set rancherImage=${RANCHER_IMAGE}"
    VERSION="--version ${CHART_VERSION}"
else
    IMAGE_REGISTRY="${RANCHER_IMAGE%%/*}"

    if [[ -n "$RANCHER_AGENT_IMAGE" ]]; then
        IMAGE_REPOSITORY="rancher/rancher"
    else
        IMAGE_REPOSITORY="${RANCHER_IMAGE#*/}"
    fi
    
    RANCHER_TAG="--set image.tag=${RANCHER_TAG_VERSION}"
    IMAGE="--set image.repository=${IMAGE_REPOSITORY} --set image.registry=${IMAGE_REGISTRY}"
    VERSION="--version ${CHART_VERSION}"
fi

set -ex

install_kubectl() {
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
}

check_cluster_status() {
    EXPECTED_NODES=3
    TIMEOUT=300
    INTERVAL=10
    ELAPSED=0

    while true; do
        TOTAL_NODES=$(kubectl get nodes --no-headers | wc -l)
        READY_NODES=$(kubectl get nodes --no-headers | awk '$2 == "Ready"' | wc -l)

        if [ "$READY_NODES" -ne "$EXPECTED_NODES" ]; then
            echo "Waiting for all $EXPECTED_NODES nodes to be Ready...$READY_NODES/$TOTAL_NODES Ready"
            sleep $INTERVAL

            ELAPSED=$((ELAPSED + INTERVAL))

            if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
                echo "Timeout reached: Not all nodes are Ready after $TIMEOUT seconds."
                exit 1
            fi
        else
            echo "All nodes are in status Ready!"
            break
        fi
    done
}

install_helm() {
    echo "Installing Helm"
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    chmod +x get_helm.sh
    ./get_helm.sh
    rm get_helm.sh
}

setup_helm_repo() {
    echo "Adding Helm chart repo"
    helm repo add rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
}

install_cert_manager() {
    echo "Installing cert manager"
    kubectl create ns cattle-system
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.crds.yaml
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace \
                                                                                    --version ${CERT_MANAGER_VERSION} \
                                                                                    --set http_proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                    --set https_proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                    --set no_proxy="${NO_PROXY}"

    kubectl get pods --namespace cert-manager

    echo "Waiting 1 minute for Rancher"
    sleep 60
}

install_self_signed_rancher() {
    echo "Installing self-signed Rancher"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
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
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel

    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${VERSION} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                        --set noProxy="${NO_PROXY}" \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel
    fi
}

install_lets_encrypt_rancher() {
    echo "Installing Let's Encrypt Rancher"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     ${VERSION} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set 'extraEnv[0].name=RANCHER_VERSION_TYPE' \
                                                                                     --set 'extraEnv[0].value=prime' \
                                                                                     --set 'extraEnv[1].name=CATTLE_BASE_UI_BRAND' \
                                                                                     --set 'extraEnv[1].value=suse' \
                                                                                     --set 'extraEnv[2].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[2].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                     --set noProxy="${NO_PROXY}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     ${VERSION} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                     --set noProxy="${NO_PROXY}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --devel
    fi
}

wait_for_rollout() {
    echo "Waiting for Rancher to be rolled out"
    kubectl -n cattle-system rollout status deploy/rancher
    kubectl -n cattle-system get deploy rancher
}

wait_for_rancher() {
    echo "Waiting 3 minutes for Rancher to be ready to deploy downstream clusters"
    sleep 180
}

install_kubectl
check_cluster_status
install_helm
setup_helm_repo

# Needed to get the latest chart version if RANCHER_TAG_VERSION is head
if [[ $RANCHER_TAG_VERSION == head ]]; then
    LATEST_CHART_VERSION=$(helm search repo rancher-${REPO} --devel | tail -n +2 | head -n 1 | cut -f2)
    VERSION="--version ${LATEST_CHART_VERSION}"
fi

install_cert_manager

case "$CERT_TYPE" in
    "self-signed")
        install_self_signed_rancher
        ;;
    "lets-encrypt")
        install_lets_encrypt_rancher
        ;;
      *)
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
        ;;
esac

wait_for_rollout
wait_for_rancher