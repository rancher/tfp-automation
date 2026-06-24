#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_MANAGER_VERSION=$3
HOSTNAME=$4
RANCHER_TAG_VERSION=$5
CHART_VERSION=$6
BOOTSTRAP_PASSWORD=$7
RANCHER_IMAGE=$8
FULL_CHAIN_FILE=$9
CERT_KEY_FILE=${10}
RANCHER_AGENT_IMAGE=${11}
TURTLES=${12}
MCM=${13}
PARTNER_RC=${14}

USER=$(whoami)

echo "Decoding certificate files..."
base64 -d <<< "$FULL_CHAIN_FILE" > /home/$USER/fullchain.pem
base64 -d <<< "$CERT_KEY_FILE" > /home/$USER/privkey.pem

chmod 600 /home/$USER/fullchain.pem
chmod 600 /home/$USER/privkey.pem

FULL_CHAIN_PATH=/home/$USER/fullchain.pem
CERT_KEY_PATH=/home/$USER/privkey.pem

mv $FULL_CHAIN_PATH /home/$USER/tls.crt
mv $CERT_KEY_PATH /home/$USER/tls.key

if [[ $RANCHER_TAG_VERSION == v2.11* || $RANCHER_TAG_VERSION == v2.10* ]]; then
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
    KUBECTL_VERSION="v1.36.0"
    curl -fsSL --max-time 30 -o kubectl https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl
    curl -fsSL --max-time 30 -o kubectl.sha256 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256
    echo "$(cat kubectl.sha256) kubectl" | sha256sum -c
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
    HELM_VERSION="v4.1.4"
    ARCH=$(uname -m)
    if [[ $ARCH == "x86_64" ]]; then
        ARCH="amd64"
    elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
        ARCH="arm64"
    fi

    echo "Installing Helm"
    curl -fsSL --max-time 30 -o helm-${HELM_VERSION}-linux-${ARCH}.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz
    curl -fsSL --max-time 30 -o helm-${HELM_VERSION}-linux-${ARCH}.tar.gz.sha256 https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz.sha256

    echo "$(cat helm-${HELM_VERSION}-linux-${ARCH}.tar.gz.sha256) helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" | sha256sum -c
    tar -xf helm-${HELM_VERSION}-linux-${ARCH}.tar.gz
    sudo mv linux-${ARCH}/helm /usr/local/bin/helm
    rm -rf linux-${ARCH} helm-${HELM_VERSION}-linux-${ARCH}.tar.gz
    rm helm-${HELM_VERSION}-linux-${ARCH}.tar.gz.sha256
}

setup_helm_repo() {
    echo "Adding Helm chart repo"
    if [ "$REPO" == "rc" ]; then
        helm repo add rancher-partner-${REPO} ${RANCHER_CHART_REPO}${REPO}
    else
        helm repo add rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
    fi
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

install_turtles_off() {
    echo "Installing Rancher with Turtles off"
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
                                                                                            --set 'extraEnv[2].name=CATTLE_FEATURES' \
                                                                                            --set 'extraEnv[2].value=turtles=false\,embedded-cluster-api=true' \
                                                                                            --set 'extraEnv[3].name=CATTLE_AGENT_IMAGE' \
                                                                                            --set "extraEnv[3].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                            --set agentTLSMode=system-store \
                                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                            --set ingress.tls.source=secret \
                                                                                            --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${VERSION} \
                                                                                         ${RANCHER_TAG} \
                                                                                         ${IMAGE} \
                                                                                         --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                         --set 'extraEnv[0].value=turtles=false\,embedded-cluster-api=true' \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
    fi
}

install_mcm_off() {
    echo "Installing Rancher with MCM off"
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
                                                                                         --set 'extraEnv[2].name=CATTLE_FEATURES' \
                                                                                         --set 'extraEnv[2].value=multi-cluster-management=false' \
                                                                                         --set 'extraEnv[3].name=CATTLE_AGENT_IMAGE' \
                                                                                         --set "extraEnv[3].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${VERSION} \
                                                                                         ${RANCHER_TAG} \
                                                                                         ${IMAGE} \
                                                                                         --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                         --set 'extraEnv[0].value=multi-cluster-management=false' \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
    fi
}

install_default_rancher() {
    kubectl -n cattle-system create secret tls tls-rancher-ingress --cert=/home/$USER/tls.crt --key=/home/$USER/tls.key

    echo "Installing Rancher"
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
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                         --set hostname=${HOSTNAME} \
                                                                                         ${VERSION} \
                                                                                         ${RANCHER_TAG} \
                                                                                         ${IMAGE} \
                                                                                         --set agentTLSMode=system-store \
                                                                                         --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                         --set ingress.tls.source=secret \
                                                                                         --devel
    fi
}

install_partner_rc_rancher() {
    echo "Creating Rancher..."
    helm upgrade --install rancher rancher-partner-rc/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                            --set hostname=${HOSTNAME} \
                                                                            ${VERSION} \
                                                                            --set rancherImageTag="${RANCHER_TAG_VERSION}" \
                                                                            --set rancherImage="${RANCHER_IMAGE}" \
                                                                            --set 'extraEnv[0].name=RANCHER_VERSION_TYPE' \
                                                                            --set 'extraEnv[0].value=prime' \
                                                                            --set 'extraEnv[1].name=CATTLE_BASE_UI_BRAND' \
                                                                            --set 'extraEnv[1].value=suse' \
                                                                            --set 'extraEnv[2].name=CATTLE_AGENT_IMAGE' \
                                                                            --set "extraEnv[2].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                            --set agentTLSMode=system-store \
                                                                            --set systemDefaultRegistry="registry.rancher.com/rancher/rc" \
                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                            --set ingress.tls.source=secret \
                                                                            --devel
}

wait_for_rollout() {
    echo "Waiting for Rancher to be rolled out"
    kubectl -n cattle-system rollout status deploy/rancher
    kubectl -n cattle-system get deploy rancher
}

ipv6_cordns_update() {
    if kubectl get svc -n kube-system | awk 'NR>1 {print $3}' | grep -q ':'; then
        . /etc/os-release

        [[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt -y install jq
        [[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install jq -y
        [[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install  -y jq

        echo "Updating CoreDNS configmap for IPv6"
        kubectl -n kube-system get cm rke2-coredns-rke2-coredns -o json  \
        | jq '.data.Corefile |= sub("forward\\s+\\.\\s+/etc/resolv\\.conf"; "forward  . 2001:4860:4860::8888")' \
        | kubectl apply -f -
        
        kubectl -n kube-system delete pod -l k8s-app=kube-dns
    fi
}

wait_for_rancher() {
    echo "Waiting 3 minutes for Rancher to be ready to deploy downstream clusters"
    sleep 180
}

install_kubectl
check_cluster_status
install_helm
setup_helm_repo

# Needed to get the latest chart version if RANCHER_TAG_VERSION contains "head"
if [[ $RANCHER_TAG_VERSION == *head* ]]; then
    LATEST_CHART_VERSION=$(helm search repo rancher-${REPO} --devel | tail -n +2 | head -n 1 | cut -f2)
    VERSION="--version ${LATEST_CHART_VERSION}"
fi

install_cert_manager

if [ -n "$TURTLES" ]; then
    case "$TURTLES" in
        "false"|"toggledOn")
            install_turtles_off
            ;;
        *)
            install_default_rancher
            ;;
    esac
elif [ -n "$MCM" ]; then
    case "$MCM" in
        "false"|"toggledOn")
            install_mcm_off
            ;;
        *)
            install_default_rancher
            ;;
    esac
elif [ -n "$PARTNER_RC" ]; then
    install_partner_rc_rancher
else
    install_default_rancher
fi

wait_for_rollout
ipv6_cordns_update
wait_for_rancher