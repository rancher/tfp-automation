#!/bin/bash

USER=$1
GROUP=$2
BASTION=$3
PASSWORD=$4
REGISTRY_USER=$5
REGISTRY_PASS=$6
REGISTRY_NAME=$7
K8S_VERSION=$8
RKE2_SERVER_ONE_IP=$9
RKE2_SERVER_TWO_IP=${10}
RKE2_SERVER_THREE_IP=${11}
DOCKER_DIR="/etc/systemd/system/docker.service.d"
PORT="3228"

set -e

echo "Setting up htpasswd..."
. /etc/os-release

[[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt update && sudo apt install -y apache2-utils wget
[[ "${ID}" == "rhel" || "${ID}" == "fedora" ]] && sudo yum install -y httpd-tools wget
[[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install -y apache2-utils wget

if [ "$(sudo docker ps -q -f name=${REGISTRY_NAME})" ]; then
    echo "Private registry ${REGISTRY_NAME} already exists. Skipping..."
else
    sudo mkdir -p /home/${USER}/auth
    sudo htpasswd -Bbn ${REGISTRY_USER} ${REGISTRY_PASS} | sudo tee /home/${USER}/auth/htpasswd

    echo "Creating a self-signed certificate..."
    sudo mkdir -p /home/${USER}/certs
    sudo openssl req -newkey rsa:4096 -nodes -sha256 -keyout /home/${USER}/certs/domain.key -addext "subjectAltName = DNS:${BASTION}" -x509 -days 365 -out /home/${USER}/certs/domain.crt -subj "/C=US/ST=CA/L=SUSE/O=Dis/CN=${BASTION}"

    echo "Copying the certificate to the /etc/docker/certs.d/${BASTION} directory..."
    sudo mkdir -p /etc/docker/certs.d/${BASTION}
    sudo cp /home/${USER}/certs/domain.crt /etc/docker/certs.d/${BASTION}/ca.crt

    echo "Creating a private registry..."
    sudo docker run -d --restart=always --name "${REGISTRY_NAME}" -v /home/${USER}/auth:/auth -v /home/${USER}/certs:/certs \
                                                                                                    -e REGISTRY_AUTH=htpasswd \
                                                                                                    -e REGISTRY_AUTH_HTPASSWD_REALM="Registry Realm" \
                                                                                                    -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
                                                                                                    -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
                                                                                                    -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
                                                                                                    -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
                                                                                                    -p 443:443 \
                                                                                                    registry:2

    echo "Logging into the private registry..."
    sudo docker login https://${BASTION} -u ${REGISTRY_USER} -p ${REGISTRY_PASS}
fi

echo "Starting proxy..."
sudo mkdir -p /home/$USER/squid
PROXY_DIR=/home/$USER/squid
sudo mv /tmp/squid.conf ${PROXY_DIR}/squid.conf

sudo mkdir -p /var/cache/squid
sudo chown -R ${USER}:${GROUP} /var/cache/squid
sudo chmod 777 /var/cache/squid

sudo docker run -d -v ${PROXY_DIR}/squid.conf:/etc/squid/squid.conf -v /var/cache/squid:/var/cache/squid -p ${PORT}:${PORT} ubuntu/squid

sudo mv /tmp/keyfile.pem /home/$USER/keyfile.pem
PEM=/home/$USER/keyfile.pem
sudo chown $USER:$GROUP $PEM
chmod 600 $PEM

wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2.linux-amd64.tar.gz
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2.linux-arm64.tar.gz
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2-images.linux-amd64.tar.zst
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/rke2-images.linux-arm64.tar.zst
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/sha256sum-amd64.txt
wget https://github.com/rancher/rke2/releases/download/${K8S_VERSION}%2Brke2r1/sha256sum-arm64.txt

curl -sfL https://get.rke2.io --output install.sh
chmod +x install.sh

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    KUBECTL_ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    KUBECTL_ARCH="arm64"
fi

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${KUBECTL_ARCH}/kubectl"
sudo chmod +x kubectl
sudo mv kubectl /usr/local/bin/

echo "Copying files to RKE2 server one"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null /usr/local/bin/kubectl ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_ONE_IP}:/home/${USER}/

echo "Copying files to RKE2 server two"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_TWO_IP}:/home/${USER}/

echo "Copying files to RKE2 server three"
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-amd64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2.linux-arm64.tar.gz ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-amd64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null rke2-images.linux-arm64.tar.zst ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null install.sh ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-amd64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/
sudo scp -i ${PEM} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sha256sum-arm64.txt ${USER}@${RKE2_SERVER_THREE_IP}:/home/${USER}/