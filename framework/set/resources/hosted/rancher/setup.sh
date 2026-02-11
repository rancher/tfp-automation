#!/bin/bash

RESOURCE_PREFIX=$1
PROVIDER=$2
RANCHER_CHART_REPO=$3
REPO=$4
CERT_MANAGER_VERSION=$5
HOSTNAME=$6
RANCHER_TAG_VERSION=$7
CHART_VERSION=$8
BOOTSTRAP_PASSWORD=$9
RANCHER_IMAGE=${10}
RANCHER_AGENT_IMAGE=${11}

RESOURCE_GROUP_NAME="${RESOURCE_PREFIX}-rg"

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
    TIMEOUT=600
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
    helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version ${CERT_MANAGER_VERSION}
    kubectl get pods --namespace cert-manager

    echo "Waiting 1 minute for Rancher"
    sleep 60
}

install_ingress_nginx() {
    echo "Installing ingress-nginx..."
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
    helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
                 --namespace ingress-nginx \
                 --set controller.service.type=LoadBalancer \
                 --set controller.service.annotations."service\.beta\.kubernetes\.io/azure-load-balancer-health-probe-request-path"=/healthz \
                 --set controller.service.externalTrafficPolicy=Cluster \
                 --create-namespace

    kubectl get pods --namespace ingress-nginx

    echo "Waiting for ingress-nginx to be rolled out"
    kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller

    echo "Waiting for ingress-nginx to be rolled out"
    kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller

    # if provider is aks, we need to wait for the public IP to be provisioned before proceeding, otherwise the Rancher installation will fail since it needs to set the hostname to the public IP of the ingress controller
    if [[ $PROVIDER == "aks" ]]; then
        echo "Waiting for ingress-nginx to be provisioned with a public IP..."
        while true; do
            INGRESS=$(kubectl get service ingress-nginx-controller --namespace=ingress-nginx -o wide | awk 'NR==2 {print $4}')
            if [[ -n "$INGRESS" && "$INGRESS" != "<pending>" ]]; then
                echo "Ingress-nginx is provisioned with public IP: $INGRESS"

                MC_RG=$(az aks show --resource-group "$RESOURCE_GROUP_NAME" --name "$RESOURCE_PREFIX" --query nodeResourceGroup -o tsv)

                # This will create the DNS label for the AKS public IP. This will be used to access Rancher once it is
                # later deployed.
                PUBLIC_IP_NAME=$(az network public-ip list --resource-group "$MC_RG" --query "[?starts_with(name, 'kubernetes')].name | [0]" -o tsv)

                echo "Updating public IP $PUBLIC_IP_NAME with DNS label $RESOURCE_PREFIX..."
                az network public-ip update --resource-group "$MC_RG" --name "$PUBLIC_IP_NAME" --dns-name "$RESOURCE_PREFIX" > /dev/null

                break
            else
                echo "Waiting for ingress-nginx to be provisioned with a public IP..."
                sleep 5
            fi
        done
    fi
}

install_default_rancher() {
    echo "Installing Rancher"
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${RANCHER_TAG} \
                                                                                         ${IMAGE} \
                                                                                         --set 'extraEnv[0].name=RANCHER_VERSION_TYPE' \
                                                                                         --set 'extraEnv[0].value=prime' \
                                                                                         --set 'extraEnv[1].name=CATTLE_BASE_UI_BRAND' \
                                                                                         --set 'extraEnv[1].value=suse' \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.ingressClassName=nginx \
                                                                                         --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${VERSION} \
                                                                                         ${RANCHER_TAG} \
                                                                                         ${IMAGE} \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.ingressClassName=nginx \
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
install_ingress_nginx
install_default_rancher
wait_for_rollout
wait_for_rancher