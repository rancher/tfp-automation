#!/usr/bin/bash

REGISTRY_NAME=$1
REGISTRY_USERNAME=$2
REGISTRY_PASSWORD=$3
HOST=$4
RANCHER_VERSION=$5
ASSET_DIR=$6
USER=$7
RANCHER_IMAGE=$8
RANCHER_AGENT_IMAGE=${9}

set -e

manageImages() {
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
        sudo docker pull ${IMAGE} && sudo docker tag ${IMAGE} ${HOST}/${IMAGE} &
    elif [ "$ACTION" == "push" ]; then
        sudo docker push ${HOST}/${IMAGE} &
    fi
}

copyImagesWithCrane() {
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

    declare -A IMAGE_PATTERNS=(
        ["mirrored-pause"]="mirrored-pause"
        ["system-agent-installer-rke2"]="system-agent-installer-rke2"
        ["rke2-runtime"]="rke2-runtime"
    )

    for PATTERN in "${!IMAGE_PATTERNS[@]}"; do
        mapfile -t VERSIONS < <(grep -oP "${PATTERN}:\K[^ ]+" /home/${USER}/rancher-images.txt | tail -n 30)
        for VERSION in "${VERSIONS[@]}"; do
            SRC_IMAGE="docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}"
            DEST_IMAGE="${HOST}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}"

            crane copy "${SRC_IMAGE}" "${DEST_IMAGE}" --insecure --platform all

            if [ "${PATTERN}" == "rke2-runtime" ]; then
                WINS_SUFFIX="-windows-amd64"
                SRC_IMAGE="docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}${WINS_SUFFIX}"
                DEST_IMAGE="${HOST}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}${WINS_SUFFIX}"

                crane copy "${SRC_IMAGE}" "${DEST_IMAGE}" --insecure --platform all
            fi
        done
    done

    mapfile -t WINDOWS_IMAGES < /home/${USER}/rancher-windows-images.txt
    for IMAGE in "${WINDOWS_IMAGES[@]}"; do
        crane copy "docker.io/${IMAGE}" "${HOST}/${IMAGE}" --insecure --platform all
    done
}

echo "Logging into the private registry..."
sudo docker login https://registry-1.docker.io -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD}

echo "Checking if the private registry already exists..."
if [ "$(sudo docker ps -q -f name=${REGISTRY_NAME})" ]; then
    echo "Private registry ${REGISTRY_NAME} already exists. Skipping..."
else
    echo "Creating a self-signed certificate..."
    sudo mkdir -p /home/${USER}/certs
    sudo openssl req -newkey rsa:4096 -nodes -sha256 -keyout /home/${USER}/certs/domain.key -addext "subjectAltName = DNS:${HOST}" -x509 -days 365 -out /home/${USER}/certs/domain.crt -subj "/C=US/ST=CA/L=SUSE/O=Dis/CN=${HOST}"

    echo "Copying the certificate to the /etc/docker/certs.d/${HOST} directory..."
    sudo mkdir -p /etc/docker/certs.d/${HOST}
    sudo cp /home/${USER}/certs/domain.crt /etc/docker/certs.d/${HOST}/ca.crt

    echo "Creating a private registry..."
    sudo docker run -d --restart=always --name "${REGISTRY_NAME}" -v /home/${USER}/certs:/certs \
                                                                  -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
                                                                  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
                                                                  -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
                                                                  -p 443:443 \
                                                                  registry:2
fi

if [ -f /home/${USER}/rancher-images.txt ]; then
    sudo rm -f /home/${USER}/rancher-*
fi

sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt -O /home/${USER}/rancher-images.txt
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-windows-images.txt -O /home/${USER}/rancher-windows-images.txt
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-save-images.sh -O /home/${USER}/rancher-save-images.sh
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-load-images.sh -O /home/${USER}/rancher-load-images.sh
    
sudo chmod +x /home/${USER}/rancher-save-images.sh && sudo chmod +x /home/${USER}/rancher-load-images.sh
sudo sed -i "s/docker save/# docker save /g" /home/${USER}/rancher-save-images.sh
sudo sed -i "s/docker load/# docker load /g" /home/${USER}/rancher-load-images.sh
sudo sed -i '/mirrored-prometheus-windows-exporter/d' /home/${USER}/rancher-images.txt

if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
    sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
    sudo sed -i "s|rancher/rancher-agent:|${RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
fi
    
echo "Pulling the images..."
manageImages "pull"

echo "Pushing the newly tagged images..."
manageImages "push"

echo "Copying needed Windows images with Crane..."
copyImagesWithCrane