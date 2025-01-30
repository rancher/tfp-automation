#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_MANAGER_VERSION=$3
HOSTNAME=$4
RANCHER_TAG_VERSION=$5
BOOTSTRAP_PASSWORD=$6
RANCHER_IMAGE=$7
BASTION=$8
STAGING_RANCHER_AGENT_IMAGE=${9}
PROXY_PORT="3128"
NO_PROXY="localhost\\,127.0.0.0/8\\,10.0.0.0/8\\,172.0.0.0/8\\,192.168.0.0/16\\,.svc\\,.cluster.local\\,cattle-system.svc\\,169.254.169.254"

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
helm repo add rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

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

echo "Installing Rancher"
if [ -n "$STAGING_RANCHER_AGENT_IMAGE" ]; then
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set rancherImage=${RANCHER_IMAGE} \
                                                                                 --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                 --set "extraEnv[0].value=${STAGING_RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                 --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                 --set noProxy="${NO_PROXY}" \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD} --devel

else
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImage=${RANCHER_IMAGE} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set proxy="http://${BASTION}:${PROXY_PORT}" \
                                                                                 --set noProxy="${NO_PROXY}" \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD}
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher

echo "Waiting 3 minutes for Rancher to be ready to deploy downstream clusters"
sleep 180