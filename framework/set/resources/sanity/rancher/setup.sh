#!/bin/bash

RANCHER_REPO=$1
RANCHER_CHART_REPO=$2
TYPE=$3
CERT_MANAGER_VERSION=$4
HOSTNAME=$5
RANCHER_TAG_VERSION=$6
BOOTSTRAP_PASSWORD=$7
STAGING_RANCHER_IMAGE=${8}
STAGING_RANCHER_AGENT_IMAGE=${9}

set -ex

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
mkdir -p ~/.kube
rm kubectl

echo "Installing Helm"
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod +x get_helm.sh
./get_helm.sh
rm get_helm.sh

echo "Adding Helm chart repo"
helm repo add ${RANCHER_REPO} ${RANCHER_CHART_REPO}${TYPE}

echo "Installing cert manager"
kubectl create ns cattle-system
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.crds.yaml
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version ${CERT_MANAGER_VERSION}
kubectl get pods --namespace cert-manager

echo "Waiting 1 minute for Rancher"
sleep 60

echo "Installing Rancher"
if [ -z "$STAGING_RANCHER_IMAGE" ]; then
    helm upgrade --install rancher ${RANCHER_REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD}
else
    helm upgrade --install rancher ${RANCHER_REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set rancherImage=${STAGING_RANCHER_IMAGE} \
                                                                                 --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                 --set "extraEnv[0].value=${STAGING_RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD} --devel
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher

echo "Waiting 3 minutes for Rancher to be ready to deploy downstream clusters"
sleep 180