#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
HOSTNAME=$3
INTERNAL_FQDN=$4
RANCHER_TAG_VERSION=$5
BOOTSTRAP_PASSWORD=$6
RANCHER_IMAGE=$7
REGISTRY=$8
RANCHER_AGENT_IMAGE=${9}

set -ex

echo "Adding Helm chart repo"
helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Upgrading Rancher"
if [ -n "$RANCHER_AGENT_IMAGE" ]; then
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                 --set systemDefaultRegistry=${REGISTRY} \
                                                                                 --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                 --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD} --devel

else
    helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                 --set hostname=${HOSTNAME} \
                                                                                 --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                 --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                 --set systemDefaultRegistry=${REGISTRY} \
                                                                                 --set bootstrapPassword=${BOOTSTRAP_PASSWORD} --devel
fi

echo "Waiting for Rancher to be rolled out"
kubectl -n cattle-system rollout status deploy/rancher
kubectl -n cattle-system get deploy rancher

kubectl patch ingress rancher -n cattle-system --type=json -p="[{
  \"op\": \"add\", 
  \"path\": \"/spec/rules/-\", 
  \"value\": {
    \"host\": \"${INTERNAL_FQDN}\", 
    \"http\": {
      \"paths\": [{
        \"backend\": {
          \"service\": {
            \"name\": \"rancher\",
            \"port\": {
              \"number\": 80
            }
          }
        },
        \"pathType\": \"ImplementationSpecific\"
      }]
    }
  }
}]"

kubectl patch ingress rancher -n cattle-system --type=json -p="[{
  \"op\": \"add\", 
  \"path\": \"/spec/tls/0/hosts/-\", 
  \"value\": \"${INTERNAL_FQDN}\"
}]"

kubectl patch setting server-url --type=json -p="[{
  \"op\": \"add\", 
  \"path\": \"/value\", 
  \"value\": \"https://${INTERNAL_FQDN}\"
}]"