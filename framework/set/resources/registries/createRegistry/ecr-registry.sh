#!/usr/bin/bash

ECR=$1
RANCHER_VERSION=$2
RANCHER_IMAGE=$3
USER=$4
ASSET_DIR=$5
RANCHER_AGENT_IMAGE=${6}

set -e

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
    docker pull ${RANCHER_IMAGE}
    docker pull ${RANCHER_AGENT_IMAGE}
fi

echo "Tagging the images..."
for IMAGE in $(cat rancher-images.txt); do
    docker tag ${IMAGE} ${ECR}/${IMAGE}
done

if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
    docker tag ${RANCHER_IMAGE} ${ECR}/${RANCHER_IMAGE}
    docker tag ${RANCHER_AGENT_IMAGE} ${ECR}/${RANCHER_AGENT_IMAGE}
fi

echo "Pushing the newly tagged images ECR..."
for IMAGE in $(cat rancher-images.txt); do
    docker push ${ECR}/${IMAGE}
done

if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
    docker push ${ECR}/${RANCHER_IMAGE}
    docker push ${ECR}/${RANCHER_AGENT_IMAGE}
fi