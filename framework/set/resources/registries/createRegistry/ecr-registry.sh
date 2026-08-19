#!/usr/bin/bash

ECR=$1
DOCKERHUB_USERNAME=$2
DOCKERHUB_PASSWORD=$3
RANCHER_VERSION=$4
RANCHER_IMAGE=$5
USER=$6
ASSET_DIR=$7
AWS_ACCESS_KEY_ID=$8
AWS_SECRET_ACCESS_KEY=$9
AWS_REGION=${10}
REPO=${11}
RANCHER_CHART_REPO=${12}
RANCHER_AGENT_IMAGE=${13}

set -e

PARALLEL_ACTIONS=10

configure_aws() {
    echo "Configuring AWS CLI..."
    aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
    aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"
    aws configure set default.region "$AWS_REGION"

    echo "Logging into Amazon ECR..."
    aws ecr get-login-password --region "$AWS_REGION" | sudo docker login --username AWS --password-stdin "${ECR}"
}

docker_login() {
    echo "Logging into Docker Hub..."
    sudo docker login https://registry-1.docker.io -u ${DOCKERHUB_USERNAME} -p ${DOCKERHUB_PASSWORD}
}

copy_certs() {
    echo "Creating a self-signed certificate..."
    mkdir -p certs
    openssl req -newkey rsa:4096 -nodes -sha256 -keyout certs/domain.key -addext "subjectAltName = DNS:${ECR}" \
                                                                         -x509 -days 365 -out certs/domain.crt \
                                                                         -subj "/C=US/ST=CA/L=SUSE/O=Dis/CN=${ECR}"

    echo "Copying the certificate to the /etc/docker/certs.d/${ECR} directory..."
    sudo mkdir -p /etc/docker/certs.d/${ECR}
    sudo cp certs/domain.crt /etc/docker/certs.d/${ECR}/ca.crt
}

fetch_images() {
    echo "Downloading ${RANCHER_VERSION} image list and scripts..."
    if [[ $REPO == "prime-release" ]]; then
        . /etc/os-release

        [[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt -y install yq
        [[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install yq -y
        [[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install  -y yq

        HEAD_SERIES=""
        if [[ $RANCHER_VERSION =~ ^v([0-9]+\.[0-9]+)-head$ ]]; then
            HEAD_SERIES="${BASH_REMATCH[1]}"
        fi

        if [[ -z "$HEAD_SERIES" && $RANCHER_VERSION == "head" ]]; then
            HEAD_SERIES="${RANCHER_HEAD_SERIES:-}"
            if [[ -z "$HEAD_SERIES" ]]; then
                HEAD_SERIES=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher[].version' | sed -nE 's/^([0-9]+\.[0-9]+)\.[0-9]+.*-head$/\1/p' | sort -Vu | tail -n 1)
            fi
        fi

        if [[ $RANCHER_VERSION == "head" && -z "$HEAD_SERIES" ]]; then
            LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r '.entries.rancher | map(select(.version | test("^"))) | sort_by(.created) | .[-1].version')
        fi

        if [[ -n "$HEAD_SERIES" ]]; then
            SERIES_FILTER="^${HEAD_SERIES//./\\\\.}"
            LATEST_CHART_VERSION=$(curl -fsSL ${RANCHER_CHART_REPO}/index.yaml | yq -r ".entries.rancher | map(select(.version | test(\"${SERIES_FILTER}\"))) | sort_by(.created) | .[-1].version")
        fi

        curl -fsSL --max-time 30 -o /home/${USER}/rancher-images.txt ${ASSET_DIR}v${LATEST_CHART_VERSION}/rancher-images.txt
        curl -fsSL --max-time 30 -o /home/${USER}/rancher-windows-images.txt ${ASSET_DIR}v${LATEST_CHART_VERSION}/rancher-windows-images.txt
        curl -fsSL --max-time 30 -o /home/${USER}/sha256sum.txt ${ASSET_DIR}v${LATEST_CHART_VERSION}/sha256sum.txt
    else
        curl -fsSL --max-time 30 -o /home/${USER}/rancher-images.txt ${ASSET_DIR}${RANCHER_VERSION}/rancher-images.txt
        curl -fsSL --max-time 30 -o /home/${USER}/sha256sum.txt ${ASSET_DIR}${RANCHER_VERSION}/sha256sum.txt
    fi

    echo "Validating checksums for Rancher image lists..."
    CHECKSUM_LINE=$(grep "rancher-images.txt" /home/${USER}/sha256sum.txt)
    if [ -z "$CHECKSUM_LINE" ]; then
        echo "ERROR: Checksum for rancher-images.txt not found in sha256sum.txt file!"
        exit 1
    fi

    CHECKSUM=$(echo "$CHECKSUM_LINE" | awk "{print \$1}")
    echo "$CHECKSUM  /home/${USER}/rancher-images.txt" | sha256sum -c -

    echo "Cutting the tags from the image names..."
    while read -r LINE; do
        echo ${LINE} | cut -d: -f1
    done < /home/${USER}/rancher-images.txt > /home/${USER}/rancher-images-no-tags.txt

    create_ecr_repositories

    if [ ! -z "${RANCHER_AGENT_IMAGE}" ]; then
        sudo sed -i "s|rancher/rancher:|${RANCHER_IMAGE}:|g" /home/${USER}/rancher-images.txt
        sudo sed -i "s|rancher/rancher-agent:|${RANCHER_AGENT_IMAGE}:|g" /home/${USER}/rancher-images.txt
    fi
}

create_ecr_repositories() {
    echo "Creating ECR repositories..."
    mapfile -t IMAGES < <(sort -u /home/${USER}/rancher-images-no-tags.txt)

    COUNTER=0
    for IMAGE in "${IMAGES[@]}"; do
        {
            if aws ecr describe-repositories --repository-names ${IMAGE} >/dev/null 2>&1; then
                echo "Repository ${IMAGE} already exists. Skipping."
            else
                echo "Creating repository ${IMAGE}..."
                aws ecr create-repository --repository-name ${IMAGE}
            fi
        } &

        COUNTER=$((COUNTER+1))
        if (( COUNTER % PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
}

manage_images() {
    mapfile -t IMAGES < /home/${USER}/rancher-images.txt

    COUNTER=0
    for IMAGE in "${IMAGES[@]}"; do
        docker pull ${IMAGE} && docker tag ${IMAGE} ${ECR}/${IMAGE} && docker push ${ECR}/${IMAGE} &
        COUNTER=$((COUNTER+1))
        
        if (( $COUNTER % $PARALLEL_ACTIONS == 0 )); then
            wait
        fi
    done

    wait
}

verify_images() {
    echo "Verifying images in ECR registry..."

    mapfile -t IMAGES < /home/${USER}/rancher-images.txt
    COUNTER=0

    for IMAGE in "${IMAGES[@]}"; do
        {
            TARGET_IMAGE=${ECR}/${IMAGE}
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

configure_aws
docker_login
copy_certs
fetch_images
manage_images
verify_images