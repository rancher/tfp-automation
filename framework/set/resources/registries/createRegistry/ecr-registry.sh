#!/usr/bin/bash

ECR=$1
REGISTRY_USERNAME=$2
REGISTRY_PASSWORD=$3
RANCHER_VERSION=$4
RANCHER_IMAGE=$5
USER=$6
ASSET_DIR=$7
AWS_ACCESS_KEY_ID=$8
AWS_SECRET_ACCESS_KEY=$9
AWS_REGION=${10}
RANCHER_AGENT_IMAGE=${11}

set -e

configureAWS() {
    echo "Configuring AWS CLI..."
    aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
    aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"
    aws configure set default.region "$AWS_REGION"

    echo "Logging into Amazon ECR..."
    aws ecr get-login-password --region "$AWS_REGION" | sudo docker login --username AWS --password-stdin "${ECR}"
}

manageImages() {
    ACTION=$1
    mapfile -t IMAGES < /home/${USER}/rancher-images.txt
    PARALLEL_ACTIONS=10

    COUNTER=0
    for IMAGE in "${IMAGES[@]}"; do
        action "${ACTION}" "${IMAGE}"
        COUNTER=$((COUNTER+1))
        
        if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
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

configureAWS

echo "Logging into Docker Hub..."
sudo docker login https://registry-1.docker.io -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD}

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
