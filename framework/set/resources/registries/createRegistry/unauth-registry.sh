#!/usr/bin/bash

REGISTRY_NAME=$1
CERT_MANAGER_VERSION=$2
DOCKERHUB_USER=$3
DOCKERHUB_PASSWORD=$4
HOST=$5
RANCHER_VERSION=$6
ASSET_DIR=$7
USER=$8
RANCHER_IMAGE=$9
FULL_CHAIN_FILE=${10}
CERT_KEY_FILE=${11}
ROUTE53_FQDN=${12}
RANCHER_AGENT_IMAGE=${13}

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
    sudo docker login https://registry-1.docker.io -u "${DOCKERHUB_USER}" -p "${DOCKERHUB_PASSWORD}"
}

create_registry() {
    echo "Checking if the private registry already exists..."
    if [ "$(sudo docker ps -q -f name=${REGISTRY_NAME})" ]; then
        echo "Private registry ${REGISTRY_NAME} already exists. Skipping..."
    else
        if [ -n "${ROUTE53_FQDN}" ]; then
            echo "Using provided certificates for the registry..."
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
            -v /home/${USER}/certs:/certs \
            -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
            -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
            -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
            -p 443:443 registry:2
    fi

    if [ -n "${ROUTE53_FQDN}" ]; then
        echo "Waiting for registry ${REGISTRY_ENDPOINT} to become available..."
        for i in $(seq 1 30); do
            if curl -sk --max-time 5 https://${REGISTRY_ENDPOINT}/v2/ | grep -q '{}'; then
                echo "Registry is up!"
                break
            fi
            echo "Attempt $i: registry is not ready, retrying in 10s..."
            sleep 10
        done
    fi
}

fetch_images() {
    echo "Fetching Rancher image lists..."
    if [ -f /home/${USER}/rancher-images.txt ]; then
        sudo rm -f /home/${USER}/rancher-*
    fi

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

copy_images() {
    ARCH=$(uname -m)
    if [[ $ARCH == "x86_64" ]]; then
        ARCH="x86_64"
    elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
        ARCH="arm64"
    fi

    sudo wget https://github.com/google/go-containerregistry/releases/download/v0.20.6/go-containerregistry_Linux_${ARCH}.tar.gz
    sudo tar -xf go-containerregistry_Linux_${ARCH}.tar.gz
    sudo chmod +x crane
    sudo mv crane /usr/local/bin/crane

    # In the case rancher agent image has been replaced in rancher-images,txt, it does not exist in dockerhub.
    # So we need to omit that from the source.
    if [[ -n "${RANCHER_AGENT_IMAGE}" ]]; then
        IMAGE_TAG=$(grep -m1 'rancher/rancher-agent:' /home/${USER}/rancher-images.txt | rev | cut -d: -f1 | rev)
        crane copy "${RANCHER_IMAGE}:${IMAGE_TAG}" "${REGISTRY_ENDPOINT}/${RANCHER_IMAGE}:${IMAGE_TAG}" --insecure &
        crane copy "${RANCHER_AGENT_IMAGE}:${IMAGE_TAG}" "${REGISTRY_ENDPOINT}/${RANCHER_AGENT_IMAGE}:${IMAGE_TAG}" --insecure &
    fi

    PARALLEL_ACTIONS=10
    COUNTER=0

    while read -r IMAGE; do
        [[ -z "$IMAGE" ]] && continue
        crane copy "docker.io/${IMAGE}" "${REGISTRY_ENDPOINT}/${IMAGE}" --insecure &

        COUNTER=$((COUNTER+1))
        if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done < "/home/${USER}/rancher-images.txt"

    wait

    COUNTER=0
    while read -r IMAGE; do
        [[ -z "$IMAGE" ]] && continue
        crane copy "docker.io/${IMAGE}" "${REGISTRY_ENDPOINT}/${IMAGE}" --insecure &

        COUNTER=$((COUNTER+1))
        if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done < "/home/${USER}/rancher-windows-images.txt"

    wait
}

copy_windows_images() {
    declare -A IMAGE_PATTERNS=(
        ["mirrored-pause"]="mirrored-pause"
        ["system-agent-installer-rke2"]="system-agent-installer-rke2"
        ["rke2-runtime"]="rke2-runtime"
    )

    PARALLEL_ACTIONS=10
    COUNTER=0

    for PATTERN in "${!IMAGE_PATTERNS[@]}"; do
        mapfile -t VERSIONS < <(grep -oP "${PATTERN}:\\K[^ ]+" /home/${USER}/rancher-images.txt | tail -n 30)
        for VERSION in "${VERSIONS[@]}"; do
            SRC_IMAGE="docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}"
            DEST_IMAGE="${REGISTRY_ENDPOINT}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}"

            crane copy "${SRC_IMAGE}" "${DEST_IMAGE}" --insecure --platform all &
            COUNTER=$((COUNTER+1))
            if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
                wait
            fi

            if [ "${PATTERN}" == "rke2-runtime" ]; then
                WINS_SUFFIX="-windows-amd64"
                SRC_IMAGE="docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}${WINS_SUFFIX}"
                DEST_IMAGE="${REGISTRY_ENDPOINT}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}${WINS_SUFFIX}"

                crane copy "${SRC_IMAGE}" "${DEST_IMAGE}" --insecure --platform all &

                COUNTER=$((COUNTER+1))
                if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
                    wait
                fi
            fi
        done
    done

    wait

    COUNTER=0
    mapfile -t WINDOWS_IMAGES < /home/${USER}/rancher-windows-images.txt
    for IMAGE in "${WINDOWS_IMAGES[@]}"; do
        crane copy "docker.io/${IMAGE}" "${REGISTRY_ENDPOINT}/${IMAGE}" --insecure --platform all &

        COUNTER=$((COUNTER+1))
        if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
}

verify_images() {
    echo "Verifying images in registry..."

    PARALLEL_ACTIONS=10
    COUNTER=0

    mapfile -t IMAGES < /home/${USER}/rancher-images.txt
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
}

verify_windows_images() {
    echo "Verifying Windows images in registry..."

    PARALLEL_ACTIONS=10
    COUNTER=0

    mapfile -t WINDOWS_IMAGES < /home/${USER}/rancher-windows-images.txt
    for IMAGE in "${WINDOWS_IMAGES[@]}"; do
        {
            TARGET_IMAGE=${REGISTRY_ENDPOINT}/${IMAGE}
            if sudo docker manifest inspect ${TARGET_IMAGE} >/dev/null 2>&1; then
                echo "${IMAGE} exists"
            else
                echo "${IMAGE} is missing, fixing..."
                crane copy "docker.io/${IMAGE}" "${REGISTRY_ENDPOINT}/${IMAGE}" --insecure &
                echo "${IMAGE} pushed successfully."
            fi
        } &

        COUNTER=$((COUNTER+1))
        if (( $COUNTER % $PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
}

docker_login
create_registry
fetch_images
cert_manager_images
copy_images
copy_windows_images
verify_images
verify_windows_images