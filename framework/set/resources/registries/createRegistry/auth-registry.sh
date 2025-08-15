#!/usr/bin/bash

REGISTRY_USERNAME=$1
REGISTRY_PASSWORD=$2
HOST=$3
RANCHER_VERSION=$4
ASSET_DIR=$5
USER=$6
RANCHER_IMAGE=$7
RANCHER_AGENT_IMAGE=${8}

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

echo "Logging into the private registry..."
sudo docker login https://registry-1.docker.io -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD}

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