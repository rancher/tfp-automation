#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_TYPE=$3
HOSTNAME=$4
INTERNAL_FQDN=$5
RANCHER_TAG_VERSION=$6
CHART_VERSION=$7
RANCHER_IMAGE=$8
REGISTRY=$9
RANCHER_AGENT_IMAGE=${10}

set -ex

echo "Adding Helm chart repo"
helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}

echo "Upgrading Rancher"
if [ "$CERT_TYPE" == "self-signed" ]; then
  if [ -n "$RANCHER_AGENT_IMAGE" ]; then
      helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                  --version ${CHART_VERSION} \
                                                                                  --set hostname=${HOSTNAME} \
                                                                                  --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                  --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                  --set systemDefaultRegistry=${REGISTRY} \
                                                                                  --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                  --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                  --set agentTLSMode=system-store \
                                                                                  --devel

  else
      helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                  --version ${CHART_VERSION} \
                                                                                  --set hostname=${HOSTNAME} \
                                                                                  --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                  --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                  --set systemDefaultRegistry=${REGISTRY} \
                                                                                  --set agentTLSMode=system-store \
                                                                                  --devel
  fi
elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
    RAND_STR=$(LC_ALL=C tr -dc 'a-z0-9' </dev/urandom | head -c 12)
    LETS_ENCRYPT_EMAIL="${RAND_STR}@gmail.com"

    if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --set rancherImage=${REGISTRY}/${RANCHER_IMAGE} \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
                                                                                     --set rancherImageTag=${RANCHER_TAG_VERSION} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --devel
    fi
else
    echo "Unsupported CERT_TYPE: $CERT_TYPE"
    exit 1
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