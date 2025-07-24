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
REGISTRY=${10}
RANCHER_AGENT_IMAGE=${11}
LETS_ENCRYPT_EMAIL=${12}

set -ex

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    KUBECTL_ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    KUBECTL_ARCH="arm64"
fi

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${KUBECTL_ARCH}/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
mkdir -p ~/.kube
rm kubectl

echo "Installing Helm"
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod +x get_helm.sh
./get_helm.sh
rm get_helm.sh

echo "Adding Helm chart repo"
helm repo add rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Installing cert manager"
kubectl create ns cattle-system
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.crds.yaml
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version ${CERT_MANAGER_VERSION}
kubectl get pods --namespace cert-manager

echo "Waiting 1 minute for Rancher"
sleep 60

echo "Installing Rancher with ${CERT_TYPE} certs"
if [ "$CERT_TYPE" == "self-signed" ]; then
    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                    --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                    --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                    --set systemDefaultRegistry=${REGISTRY} \
                                                                                    --set agentTLSMode=system-store \
                                                                                    --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                    --devel

    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                    --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                    --set systemDefaultRegistry=${REGISTRY} \
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
                                                                                     --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
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

echo "Waiting 3 minutes for Rancher to be ready to deploy downstream clusters"
sleep 180