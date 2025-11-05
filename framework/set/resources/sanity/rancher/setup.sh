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
RANCHER_AGENT_IMAGE=${10}
TURTLES=${11}

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

install_turtles_off() {
    echo "Installing Rancher with Turtles off"
    if [ "$CERT_TYPE" == "self-signed" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
                helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                            --set hostname=${HOSTNAME} \
                                                                                            --version ${CHART_VERSION} \
                                                                                            --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                            --set rancherImage=${RANCHER_IMAGE} \
                                                                                            --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                            --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                            --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                            --set 'extraEnv[1].value=prime' \
                                                                                            --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                            --set 'extraEnv[2].value=suse' \
                                                                                            --set 'extraEnv[3].name=CATTLE_FEATURES' \
                                                                                            --set 'extraEnv[3].value=turtles=false\,embedded-cluster-api=true' \
                                                                                            --set agentTLSMode=system-store \
                                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                            --devel
        else
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                            --set hostname=${HOSTNAME} \
                                                                                            --version ${CHART_VERSION} \
                                                                                            --set rancherImage=${RANCHER_IMAGE} \
                                                                                            --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                            --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                            --set 'extraEnv[0].value=turtles=false\,embedded-cluster-api=true' \
                                                                                            --set agentTLSMode=system-store \
                                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                            --devel
        fi
    elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                            --set hostname=${HOSTNAME} \
                                                                                            --version ${CHART_VERSION} \
                                                                                            --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                            --set rancherImage=${RANCHER_IMAGE} \
                                                                                            --set ingress.tls.source=letsEncrypt \
                                                                                            --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                            --set letsEncrypt.ingress.class=nginx \
                                                                                            --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                            --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                            --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                            --set 'extraEnv[1].value=prime' \
                                                                                            --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                            --set 'extraEnv[2].value=suse' \
                                                                                            --set 'extraEnv[3].name=CATTLE_FEATURES' \
                                                                                            --set 'extraEnv[3].value=turtles=false\,embedded-cluster-api=true' \
                                                                                            --set agentTLSMode=system-store \
                                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                            --devel
        else
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                            --set hostname=${HOSTNAME} \
                                                                                            --version ${CHART_VERSION} \
                                                                                            --set rancherImage=${RANCHER_IMAGE} \
                                                                                            --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                            --set ingress.tls.source=letsEncrypt \
                                                                                            --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                            --set letsEncrypt.ingress.class=nginx \
                                                                                            --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                            --set 'extraEnv[0].value=turtles=false\,embedded-cluster-api=true' \
                                                                                            --set agentTLSMode=system-store \
                                                                                            --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                            --devel
        fi
    else
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
    fi
}

install_default_rancher() {
    echo "Installing Rancher"
    if [ "$CERT_TYPE" == "self-signed" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                        --set rancherImage=${RANCHER_IMAGE} \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[1].value=prime' \
                                                                                        --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[2].value=suse' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel
        else
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set rancherImage=${RANCHER_IMAGE} \
                                                                                        --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel
        fi
    elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                        --set rancherImage=${RANCHER_IMAGE} \
                                                                                        --set ingress.tls.source=letsEncrypt \
                                                                                        --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[1].value=prime' \
                                                                                        --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[2].value=suse' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel
        else
            helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set rancherImage=${RANCHER_IMAGE} \
                                                                                        --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                        --set ingress.tls.source=letsEncrypt \
                                                                                        --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                        --devel
        fi
    else
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
    fi
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
else
    install_default_rancher
fi

wait_for_rollout
ipv6_cordns_update
wait_for_rancher