#!/usr/bin/bash

ECR=$1
RANCHER_VERSION=$2
RANCHER_IMAGE=$3
USER=$4
ASSET_DIR=$5
RANCHER_AGENT_IMAGE=${6}

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
        sudo docker pull ${IMAGE} && sudo docker tag ${IMAGE} ${ECR}/${IMAGE} &
    elif [ "$ACTION" == "push" ]; then
        sudo docker push ${ECR}/${IMAGE} &
    fi
}

. /etc/os-release

[[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt install -y unzip
[[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install -y unzip
[[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install -y unzip

echo "Creating a self-signed certificate..."
mkdir -p certs
openssl req -newkey rsa:4096 -nodes -sha256 -keyout certs/domain.key -addext "subjectAltName = DNS:${ECR}" -x509 -days 365 -out certs/domain.crt -subj "/C=US/ST=CA/L=SUSE/O=Dis/CN=${ECR}"

echo "Copying the certificate to the /etc/docker/certs.d/${ECR} directory..."
sudo mkdir -p /etc/docker/certs.d/${ECR}
sudo cp certs/domain.crt /etc/docker/certs.d/${ECR}/ca.crt

echo "Downloading ${RANCHER_VERSION} image list and scripts..."
wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt
wget ${ASSET_DIR}${RANCHER_VERSION}/rancher-save-images.sh

echo "Cutting the tags from the image names..."
while read LINE; do
    echo ${LINE} | cut -d: -f1
done < rancher-images.txt > rancher-images-no-tags.txt

echo "Creating ECR repositories..."
for IMAGE in $(cat rancher-images-no-tags.txt); do
    if aws ecr describe-repositories --repository-names ${IMAGE} >/dev/null 2>&1; then
        echo "Repository ${IMAGE} already exists. Skipping."
    else
        echo "Creating repository $IMAGE..."
        aws ecr create-repository --repository-name ${IMAGE}
    fi
done

echo "Saving the images..."
sudo sed -i "s/docker save/# docker save /g" /home/${USER}/rancher-save-images.sh

chmod +x rancher-save-images.sh
./rancher-save-images.sh --image-list ./rancher-images.txt

if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
    sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
    sudo sed -i "s|rancher/rancher-agent:|${RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
fi

echo "Pulling the images..."
manageImages "pull"

echo "Pushing the newly tagged images..."
manageImages "push"