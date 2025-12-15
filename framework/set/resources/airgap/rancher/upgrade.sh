#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_TYPE=$3
HOSTNAME=$4
RANCHER_TAG_VERSION=$5
CHART_VERSION=$6
RANCHER_IMAGE=$7
REGISTRY=$8
RANCHER_AGENT_IMAGE=${9}
UPGRADED_TURTLES=${10}

if [[ $RANCHER_TAG_VERSION == v2.11* ]]; then
    RANCHER_TAG="--set rancherImageTag=${RANCHER_TAG_VERSION}" 
    IMAGE="--set rancherImage=${REGISTRY}/${RANCHER_IMAGE}"
else
    IMAGE_REGISTRY="${RANCHER_IMAGE%%/*}"

    if [[ -n "$RANCHER_AGENT_IMAGE" ]]; then
        IMAGE_REPOSITORY="rancher"
    else
        IMAGE_REPOSITORY="${RANCHER_IMAGE#*/}"
    fi
    
    RANCHER_TAG="--set image.tag=${RANCHER_TAG_VERSION}"
    IMAGE="--set image.repository=${IMAGE_REPOSITORY} --set image.registry=${REGISTRY}/${IMAGE_REGISTRY}"
fi

set -ex

setup_helm_repo() {
  echo "Adding Helm chart repo"
  helm repo add upgraded-rancher-${REPO} ${RANCHER_CHART_REPO}${REPO}
}

upgrade_turtles_off() {
    echo "Upgrading Rancher with Turtles off"
    if [ "$CERT_TYPE" == "self-signed" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                        --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                        --set 'extraEnv[1].value=prime' \
                                                                                        --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                        --set 'extraEnv[2].value=suse' \
                                                                                        --set 'extraEnv[3].name=CATTLE_FEATURES' \
                                                                                        --set 'extraEnv[3].value=turtles=false\,embedded-cluster-api=true' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel

        else
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                        --set 'extraEnv[0].value=turtles=false\,embedded-cluster-api=true' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel
        fi
    elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
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
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel
        else
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set ingress.tls.source=letsEncrypt \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set 'extraEnv[0].name=CATTLE_FEATURES' \
                                                                                        --set 'extraEnv[0].value=turtles=false\,embedded-cluster-api=true' \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel
        fi
    else
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
    fi
}

upgrade_default_rancher() {
    echo "Upgrading Rancher"
    if [ "$CERT_TYPE" == "self-signed" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                        --set "extraEnv[0].value=${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
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
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel
        fi
    elif [ "$CERT_TYPE" == "lets-encrypt" ]; then
        if [ -n "$RANCHER_AGENT_IMAGE" ]; then
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
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
                                                                                        --set useBundledSystemChart=true \
                                                                                        --devel
        else
            helm upgrade --install rancher upgraded-rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                        --version ${CHART_VERSION} \
                                                                                        --set hostname=${HOSTNAME} \
                                                                                        ${RANCHER_TAG} \
                                                                                        ${IMAGE} \
                                                                                        --set systemDefaultRegistry=${REGISTRY} \
                                                                                        --set ingress.tls.source=letsEncrypt \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                        --set letsEncrypt.ingress.class=nginx \
                                                                                        --set agentTLSMode=system-store \
                                                                                        --set useBundledSystemChart=true \
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

patch_rancher_fqdn() {
  echo "Patching Rancher FQDN: ${HOSTNAME}"
  kubectl patch ingress rancher -n cattle-system --type=json -p="[{
    \"op\": \"add\", 
    \"path\": \"/spec/rules/-\", 
    \"value\": {
      \"host\": \"${HOSTNAME}\", 
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
    \"value\": \"${HOSTNAME}\"
  }]"

  kubectl patch setting server-url --type=json -p="[{
    \"op\": \"add\", 
    \"path\": \"/value\", 
    \"value\": \"https://${HOSTNAME}\"
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

if [ -n "$UPGRADED_TURTLES" ]; then
    case "$UPGRADED_TURTLES" in
        "false"|"toggledOn")
            upgrade_turtles_off
            ;;
        *)
            upgrade_default_rancher
            ;;
    esac
else
    upgrade_default_rancher
fi


wait_for_rollout
patch_rancher_fqdn
wait_for_rancher