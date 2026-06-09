#!/usr/bin/bash

CERT_MANAGER_VERSION=$1
REGISTRY_NAME=$2
REGISTRY_USER=$3
REGISTRY_PASS=$4
DOCKERHUB_USER=$5
DOCKERHUB_PASS=$6
HOST=$7
RANCHER_VERSION=$8
ASSET_DIR=$9
USER=${10}
RANCHER_IMAGE=${11}
FULL_CHAIN_FILE=${12}
CERT_KEY_FILE=${13}
ROUTE53_FQDN=${14}
RANCHER_AGENT_IMAGE=${15}

set -e

if [ ! -d "/home/$USER/certs" ]; then
    echo "Decoding certificate files..."
    base64 -d <<< "$FULL_CHAIN_FILE" > /home/$USER/fullchain.pem
    base64 -d <<< "$CERT_KEY_FILE" > /home/$USER/privkey.pem

    chmod 600 /home/$USER/fullchain.pem
    chmod 600 /home/$USER/privkey.pem
else
    echo "Certificate files already exist. Skipping decoding..."
fi

FULL_CHAIN_PATH=/home/$USER/fullchain.pem
CERT_KEY_PATH=/home/$USER/privkey.pem

REGISTRY_ENDPOINT="${HOST}"
if [ -n "${ROUTE53_FQDN}" ]; then
    REGISTRY_ENDPOINT="${ROUTE53_FQDN}"
fi

docker_login() {
    echo "Logging into Docker Hub..."
    sudo docker login https://registry-1.docker.io -u "${DOCKERHUB_USER}" -p "${DOCKERHUB_PASS}"
}

create_registry() {
    if [ "$(sudo docker ps -q -f name=${REGISTRY_NAME})" ]; then
        echo "Private registry ${REGISTRY_NAME} already exists. Skipping..."
    else
        sudo mkdir -p /home/${USER}/auth
        sudo htpasswd -Bbn ${REGISTRY_USER} ${REGISTRY_PASS} | sudo tee /home/${USER}/auth/htpasswd

        if [ -n "${ROUTE53_FQDN}" ]; then
            sudo mkdir -p /home/${USER}/certs
            sudo mv $FULL_CHAIN_PATH /home/${USER}/certs/domain.crt
            sudo mv $CERT_KEY_PATH /home/${USER}/certs/domain.key
            sudo chmod 644 /home/${USER}/certs/domain.crt
            sudo chmod 600 /home/${USER}/certs/domain.key
        else
            echo "Creating a self-signed certificate..."
            sudo mkdir -p /home/${USER}/certs
            sudo openssl req -newkey rsa:4096 -nodes -sha256 \
                -keyout /home/${USER}/certs/domain.key \
                -addext "subjectAltName = DNS:${REGISTRY_ENDPOINT}" \
                -x509 -days 365 -out /home/${USER}/certs/domain.crt \
                -subj "/C=US/ST=CA/L=SUSE/O=Dis/CN=${REGISTRY_ENDPOINT}"
        fi

        echo "Copying the certificate to the /etc/docker/certs.d/${REGISTRY_ENDPOINT} directory..."
        sudo mkdir -p /etc/docker/certs.d/${REGISTRY_ENDPOINT}
        sudo cp /home/${USER}/certs/domain.crt /etc/docker/certs.d/${REGISTRY_ENDPOINT}/ca.crt

        echo "Creating a private registry..."
        sudo docker run -d --restart=always --name "${REGISTRY_NAME}" \
            -v /home/${USER}/auth:/auth \
            -v /home/${USER}/certs:/certs \
            -e REGISTRY_AUTH=htpasswd \
            -e REGISTRY_AUTH_HTPASSWD_REALM="Registry Realm" \
            -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
            -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
            -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
            -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
            -p 443:443 registry:2
    fi

    if [ -n "${ROUTE53_FQDN}" ]; then
        echo "Waiting for registry ${REGISTRY_ENDPOINT} to become available..."
        for i in $(seq 1 30); do
            if curl -sk --max-time 5 -u "${REGISTRY_USER}:${REGISTRY_PASS}" https://${REGISTRY_ENDPOINT}/v2/ | grep -q '{}'; then
                echo "Registry is up!"
                break
            fi
            echo "Attempt $i: registry is not ready, retrying in 10s..."
            sleep 10
        done
    fi

    echo "Logging into private registry ${REGISTRY_ENDPOINT}..."
    sudo docker login -u ${REGISTRY_USER} -p ${REGISTRY_PASS} https://${REGISTRY_ENDPOINT}
}

fetch_images() {
    echo "Fetching Rancher image lists..."
    curl -fsSL --max-time 30 -o /home/${USER}/rancher-images.txt ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt
    curl -fsSL --max-time 30 -o /home/${USER}/rancher-windows-images.txt ${ASSET_DIR}${RANCHER_VERSION}/rancher-windows-images.txt
    curl -fsSL --max-time 30 -o /home/${USER}/sha256sum.txt ${ASSET_DIR}${RANCHER_VERSION}/sha256sum.txt

    echo "Validating checksums for Rancher image lists..."
    CHECKSUM_LINE=$(grep "rancher-images.txt" /home/${USER}/sha256sum.txt)
    if [ -z "$CHECKSUM_LINE" ]; then
        echo "ERROR: Checksum for rancher-images.txt not found in sha256sum.txt file!"
        exit 1
    fi

    CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
    echo "$CHECKSUM  /home/${USER}/rancher-images.txt" | sha256sum -c -

    WIN_CHECKSUM_LINE=$(grep "rancher-windows-images.txt" /home/${USER}/sha256sum.txt)
    if [ -z "$WIN_CHECKSUM_LINE" ]; then
        echo "ERROR: Checksum for rancher-windows-images.txt not found in sha256sum.txt file!"
        exit 1
    fi

    WIN_CHECKSUM=$(echo "$WIN_CHECKSUM_LINE" | awk "{print \$1}")
    echo "$WIN_CHECKSUM  /home/${USER}/rancher-windows-images.txt" | sha256sum -c -

    if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
        sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
        sudo sed -i "s|rancher/rancher-agent:|${RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
    fi
}

cert_manager_images() {
    echo "Adding cert-manager images to the list..."
    CERT_MANAGER_IMAGES=(
        "quay.io/jetstack/cert-manager-controller:${CERT_MANAGER_VERSION}"
        "quay.io/jetstack/cert-manager-webhook:${CERT_MANAGER_VERSION}"
        "quay.io/jetstack/cert-manager-cainjector:${CERT_MANAGER_VERSION}"
        "quay.io/jetstack/cert-manager-startupapicheck:${CERT_MANAGER_VERSION}"
    )

    for IMAGE in "${CERT_MANAGER_IMAGES[@]}"; do
        sudo docker pull ${IMAGE}
        sudo docker tag ${IMAGE} ${REGISTRY_ENDPOINT}/${IMAGE}
        sudo docker push ${REGISTRY_ENDPOINT}/${IMAGE}
    done
}

manage_images() {
    ACTION=$1
    mapfile -t IMAGES < /home/${USER}/rancher-images.txt
    PARALLEL_ACTIONS=10

    COUNTER=0
    for IMAGE in "${IMAGES[@]}"; do
        action "${ACTION}" "${IMAGE}"
        COUNTER=$((COUNTER+1))
        
        if (( $COUNTER % $PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
}

action() {
    ACTION=$1
    IMAGE=$2
    
    if [ "$ACTION" == "pull" ]; then
        sudo docker pull ${IMAGE} && sudo docker tag ${IMAGE} ${REGISTRY_ENDPOINT}/${IMAGE} &
    elif [ "$ACTION" == "push" ]; then
        sudo docker push ${REGISTRY_ENDPOINT}/${IMAGE} &
    fi
}

verify_images() {
    echo "Verifying images in registry..."

    mapfile -t IMAGES < /home/${USER}/rancher-images.txt
    PARALLEL_ACTIONS=10
    COUNTER=0

    for IMAGE in "${IMAGES[@]}"; do
        {
            TARGET_IMAGE=${REGISTRY_ENDPOINT}/${IMAGE}
            if sudo docker manifest inspect ${TARGET_IMAGE} >/dev/null 2>&1; then
                echo "${IMAGE} exists"
            else
                echo "${IMAGE} is missing, fixing..."
                sudo docker pull ${IMAGE}
                sudo docker tag ${IMAGE} ${TARGET_IMAGE}
                sudo docker push ${TARGET_IMAGE}
                echo "${IMAGE} pushed successfully."
            fi
        } &
        
        COUNTER=$((COUNTER+1))
        if (( $COUNTER % $PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
    echo "Image verification complete."
}

docker_login
create_registry
fetch_images
cert_manager_images
manage_images "pull"
manage_images "push"
verify_images