#!/bin/bash

RANCHER_CHART_REPO=$1
REPO=$2
CERT_MANAGER_VERSION=$3
CERT_TYPE=$4
HOSTNAME=$5
INTERNAL_FQDN=$6
RANCHER_TAG_VERSION=$7
CHART_VERSION=$8
BOOTSTRAP_PASSWORD=$9
RANCHER_IMAGE=${10}
REGISTRY=${11}
RANCHER_AGENT_IMAGE=${12}

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

check_cluster_status() {
    EXPECTED_NODES=3
    TIMEOUT=300
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

  helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version ${CERT_MANAGER_VERSION} \
                                                            --set image.repository=${REGISTRY}/quay.io/jetstack/cert-manager-controller \
                                                            --set webhook.image.repository=${REGISTRY}/quay.io/jetstack/cert-manager-webhook \
                                                            --set cainjector.image.repository=${REGISTRY}/quay.io/jetstack/cert-manager-cainjector \
                                                            --set startupapicheck.image.repository=${REGISTRY}/quay.io/jetstack/cert-manager-startupapicheck
  
  kubectl get pods --namespace cert-manager

  echo "Waiting 1 minute for Rancher"
  sleep 60
}

install_self_signed_rancher() {
  echo "Installing Rancher with self-signed certs"
  if [ -n "$RANCHER_AGENT_IMAGE" ]; then
      helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                  --set hostname=${HOSTNAME} \
                                                                                  --version ${CHART_VERSION} \
                                                                                  ${RANCHER_TAG} \
                                                                                  ${IMAGE} \
                                                                                  --set systemDefaultRegistry=${REGISTRY} \
                                                                                  --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                  --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                  --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                  --set 'extraEnv[1].value=prime' \
                                                                                  --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                  --set 'extraEnv[2].value=suse' \
                                                                                  --set agentTLSMode=system-store \
                                                                                  --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                  --set useBundledSystemChart=true \
                                                                                  --devel

  else
      helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                  --set hostname=${HOSTNAME} \
                                                                                  --version ${CHART_VERSION} \
                                                                                  ${RANCHER_TAG} \
                                                                                  ${IMAGE} \
                                                                                  --set systemDefaultRegistry=${REGISTRY} \
                                                                                  --set agentTLSMode=system-store \
                                                                                  --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                  --set useBundledSystemChart=true \
                                                                                  --devel
  fi
}

install_lets_encrypt_rancher() {
  echo "Installing Rancher with Let's Encrypt certs"
  if [ -n "$RANCHER_AGENT_IMAGE" ]; then
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --version ${CHART_VERSION} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set 'extraEnv[0].name=CATTLE_AGENT_IMAGE' \
                                                                                     --set "extraEnv[0].value=${REGISTRY}/${RANCHER_AGENT_IMAGE}:${RANCHER_TAG_VERSION}" \
                                                                                     --set 'extraEnv[1].name=RANCHER_VERSION_TYPE' \
                                                                                     --set 'extraEnv[1].value=prime' \
                                                                                     --set 'extraEnv[2].name=CATTLE_BASE_UI_BRAND' \
                                                                                     --set 'extraEnv[2].value=suse' \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --set useBundledSystemChart=true \
                                                                                     --devel
    else
        helm upgrade --install rancher rancher-${REPO}/rancher --namespace cattle-system --set global.cattle.psp.enabled=false \
                                                                                     --set hostname=${HOSTNAME} \
                                                                                     --version ${CHART_VERSION} \
                                                                                     ${RANCHER_TAG} \
                                                                                     ${IMAGE} \
                                                                                     --set systemDefaultRegistry=${REGISTRY} \
                                                                                     --set ingress.tls.source=letsEncrypt \
                                                                                     --set letsEncrypt.ingress.class=nginx \
                                                                                     --set letsEncrypt.email=${LETS_ENCRYPT_EMAIL} \
                                                                                     --set agentTLSMode=system-store \
                                                                                     --set bootstrapPassword=${BOOTSTRAP_PASSWORD} \
                                                                                     --set useBundledSystemChart=true \
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

wait_for_ingress() {
  EXPECTED_INGRESS=1
  TIMEOUT=300
  INTERVAL=10
  ELAPSED=0

  while true; do
    TOTAL_INGRESS=$(kubectl get ingress rancher -n cattle-system | awk '$1 == "rancher"' | wc -l)
    
    if [ "$TOTAL_INGRESS" -ne "$EXPECTED_INGRESS" ]; then
      echo "Waiting for $TOTAL_INGRESS Rancher ingress to be created...$ELAPSED/$TIMEOUT seconds elapsed."
      sleep $INTERVAL

      ELAPSED=$((ELAPSED + INTERVAL))
      
      if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
        echo "Timeout reached: Rancher ingress not found after $TIMEOUT seconds."
        exit 1
      fi
    else
      echo "Rancher ingress found!"
      break
    fi
  done
}

patch_rancher_internal_fqdn() {
  echo "Patching Rancher to add internal FQDN: ${INTERNAL_FQDN}"
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

check_cluster_status
install_helm
setup_helm_repo
install_cert_manager

case "$CERT_TYPE" in
    "self-signed")
        install_self_signed_rancher
        ;;
    "lets-encrypt")
        install_lets_encrypt_rancher
        ;;
      *)
        echo "Unsupported CERT_TYPE: $CERT_TYPE"
        exit 1
        ;;
esac

wait_for_rollout
wait_for_rancher
wait_for_ingress
patch_rancher_internal_fqdn