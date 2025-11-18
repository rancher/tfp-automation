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

if [[ $RANCHER_TAG_VERSION == v2.11* ]]; then
    RANCHER_TAG="--set rancherImageTag=${RANCHER_TAG_VERSION}" 
    IMAGE="--set rancherImage=${REGISTRY}/${RANCHER_IMAGE}"
else
    IMAGE_REGISTRY="${RANCHER_IMAGE%%/*}"
    IMAGE_REPOSITORY="${RANCHER_IMAGE#*/}"
    
    RANCHER_TAG="--set image.tag=${RANCHER_TAG_VERSION}"
    IMAGE="--set image.repository=${IMAGE_REPOSITORY} --set image.registry=${REGISTRY}/${IMAGE_REGISTRY}"
fi

set -ex

setup_helm_repo() {
  echo "Adding Helm chart repo"
  helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
}

upgrade_self_signed_rancher() {
  echo "Upgrading self-signed Rancher"
  if [ -n "$RANCHER_AGENT_IMAGE" ]; then
      helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    ${RANCHER_TAG} \
                                                                                    ${IMAGE} \
                                                                                    --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                    --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                    --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                    --set 'extraEnv[1].value=prime' \
                                                                                    --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                    --set 'extraEnv[2].value=suse' \
                                                                                    --set agentTLSMode=system-store \
                                                                                    --set useBundledSystemChart=true \
                                                                                    --devel

  else
      helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                    --version ${CHART_VERSION} \
                                                                                    --set hostname=${HOSTNAME} \
                                                                                    ${RANCHER_TAG} \
                                                                                    ${IMAGE} \
                                                                                    --set agentTLSMode=system-store \
                                                                                    --set useBundledSystemChart=true \
                                                                                    --devel
  fi
}

upgrade_lets_encrypt_rancher() {
  echo "Upgrading Lets Encrypt Rancher"
  if [ -n "$RANCHER_AGENT_IMAGE" ]; then
      helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                     --set 'extraEnv[1].value=prime' \
                                                                                     --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                     --set 'extraEnv[2].value=suse' \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set useBundledSystemChart=true \
                                                                                     --devel
  else
      helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --version ${CHART_VERSION} \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set useBundledSystemChart=true \
                                                                                     --devel
  fi
}

wait_for_rollout() {
  echo "Waiting for Rancher to be rolled out"
  kubectl -n cattle-system rollout status deploy/rancher
  kubectl -n cattle-system get deploy rancher
}

patch_rancher_internal_fqdn() {
  echo "Patching Rancher Ingress and Setting for internal FQDN: ${INTERNAL_FQDN}"
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

  echo "Restarting Rancher"
  kubectl -n cattle-system rollout restart deploy/rancher
  kubectl -n cattle-system rollout status deploy/rancher
  kubectl -n cattle-system get deploy rancher
}

wait_for_rancher() {
  echo "Waiting 15 seconds to be able to login to Rancher"
  sleep 15
}

setup_helm_repo

case "$CERT_TYPE" in
    "self-signed")
        upgrade_self_signed_rancher
        ;;
    "lets-encrypt")
        upgrade_lets_encrypt_rancher
        ;;
      *)
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
        ;;
esac

wait_for_rollout
patch_rancher_internal_fqdn
wait_for_rancher