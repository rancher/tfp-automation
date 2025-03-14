#!/usr/bin/bash

REGISTRY_NAME=$1
HOST=$2
RANCHER_VERSION=$3
ASSET_DIR=$4
USER=$5
RANCHER_IMAGE=$6
RANCHER_AGENT_IMAGE=${7}

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

copyImagesWithSkopeo() {
    if skopeo -v > /dev/null 2>&1; then
        echo -e "\nSkopeo is already installed"
    else
        . /etc/os-release

        [[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt -y install skopeo
        [[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install skopeo -y
        [[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install  -y skopeo
    fi

    declare -A IMAGE_PATTERNS=(
        ["system-agent-installer-rke2"]="system-agent-installer-rke2"
        ["rke2-runtime"]="rke2-runtime"
    )
    
    for PATTERN in "${!IMAGE_PATTERNS[@]}"; do
        if [ "${PATTERN}" == "rke2-runtime" ]; then
            mapfile -t VERSIONS < <(grep -oP "${PATTERN}:\K[^ ]+" /home/${USER}/rancher-images.txt | sort -rV | head -n 2)
            for VERSION in "${VERSIONS[@]}"; do
                skopeo copy -a docker://docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}-windows-amd64 docker://${HOST}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}-windows-amd64
            done
        else
            mapfile -t VERSIONS < <(grep -oP "${PATTERN}:\K[^ ]+" /home/${USER}/rancher-images.txt | sort -rV | head -n 2)
            for VERSION in "${VERSIONS[@]}"; do
                skopeo copy -a docker://docker.io/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION} docker://${HOST}/rancher/${IMAGE_PATTERNS[$PATTERN]}:${VERSION}
            done
        fi
    done

    mapfile -t WINDOWS_IMAGES < <(grep -oP "wins:\K[^ ]+" /home/${USER}/rancher-windows-images.txt)
    skopeo copy -a docker://docker.io/rancher/wins:${WINDOWS_IMAGES[0]} docker://${HOST}/rancher/wins:${WINDOWS_IMAGES[0]}
}

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

echo "Copying needed Windows images with skopeo..."
copyImagesWithSkopeo