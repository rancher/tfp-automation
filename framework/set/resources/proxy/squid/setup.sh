#!/bin/bash

USER=$1
BASTION=$2
PASSWORD=$3
DOCKER_DIR="/etc/systemd/system/docker.service.d"
PORT="3128"

set -e
   
echo "Starting proxy..."
sudo mkdir -p /home/$USER/squid
PROXY_DIR=/home/$USER/squid
sudo mv /tmp/squid.conf ${PROXY_DIR}/squid.conf

sudo docker run -d -v ${PROXY_DIR}/squid.conf:/etc/squid/squid.conf -p ${PORT}:${PORT} ubuntu/squid