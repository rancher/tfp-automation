#!/bin/bash

USER=$1
GROUP=$2
K8S_VERSION=$3
K3S_SERVER_IP=$4
K3S_NEW_SERVER_IP=$5
K3S_TOKEN=$6

set -e

sudo hostnamectl set-hostname ${K3S_NEW_SERVER_IP}

curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${K8S_VERSION} K3S_TOKEN=${K3S_TOKEN} sh -s - server --server https://${K3S_SERVER_IP}:6443