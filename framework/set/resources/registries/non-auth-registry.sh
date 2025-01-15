#!/usr/bin/bash

REGISTRY_NAME=$1
HOST=$2
RANCHER_VERSION=$3
ASSET_DIR=$4
USER=$5

HOST="${HOST}:5000"

set -e

echo "Creating a private registry..."
sudo docker run -d --restart=always --name ${REGISTRY_NAME} -p 5000:5000 registry:2

echo "Setting insecure registry in /etc/docker/daemon.json..."
sudo touch /etc/docker/daemon.json
echo "{ \"insecure-registries\": [\"${HOST}\"] }" | sudo tee /etc/docker/daemon.json
sudo systemctl restart docker && sudo systemctl daemon-reload

sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt -O /home/${USER}/rancher-images.txt
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-save-images.sh -O /home/${USER}/rancher-save-images.sh
sudo wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-load-images.sh -O /home/${USER}/rancher-load-images.sh
    
sudo chmod +x /home/${USER}/rancher-save-images.sh && sudo chmod +x /home/${USER}/rancher-load-images.sh
sudo sed -i "s/docker save/# docker save /g" /home/${USER}/rancher-save-images.sh
sudo sed -i "s/docker load/# docker load /g" /home/${USER}/rancher-load-images.sh
    
echo "Saving the images..."
sudo /home/${USER}/rancher-save-images.sh --image-list /home/${USER}/rancher-images.txt

echo "Tagging the images..."
for IMAGE in $(cat /home/$USER/rancher-images.txt); do
    sudo docker tag ${IMAGE} ${HOST}/${IMAGE}
done

echo "Pushing the newly tagged images..."
for IMAGE in $(cat /home/$USER/rancher-images.txt); do
    sudo docker push ${HOST}/${IMAGE}
done