#!/usr/bin/bash

REGISTRY_NAME=$1
HOST=$2
RANCHER_VERSION=$3
ASSET_DIR=$4
USER=$5
RANCHER_IMAGE=$6
STAGING_RANCHER_AGENT_IMAGE=${7:-""}
PRIME_RANCHER_AGENT_IMAGE=${8}

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

sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt -O /home/${USER}/rancher-images.txt
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-save-images.sh -O /home/${USER}/rancher-save-images.sh
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-load-images.sh -O /home/${USER}/rancher-load-images.sh
    
sudo chmod +x /home/${USER}/rancher-save-images.sh && sudo chmod +x /home/${USER}/rancher-load-images.sh
sudo sed -i "s/docker save/# docker save /g" /home/${USER}/rancher-save-images.sh
sudo sed -i "s/docker load/# docker load /g" /home/${USER}/rancher-load-images.sh
sudo sed -i '/mirrored-prometheus-windows-exporter/d' /home/${USER}/rancher-images.txt

if [ ! -z "${STAGING_RANCHER_AGENT_IMAGE}" ]; then
    sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
    sudo sed -i "s|rancher/rancher-agent:|${STAGING_RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
fi

if [[ ! -z "${PRIME_RANCHER_AGENT_IMAGE}" ]]; then
    sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
    sudo sed -i "s|rancher/rancher-agent:|${PRIME_RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
fi
    
echo "Pulling the images..."
manageImages "pull"

echo "Pushing the newly tagged images..."
manageImages "push"