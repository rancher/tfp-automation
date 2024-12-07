#!/bin/bash

KUBE_CONFIG=$1

set -ex

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
mkdir -p ~/.kube
rm kubectl

echo "$KUBE_CONFIG" | base64 -d > ~/.kube/config

echo "Checking nodes"
kubectl get nodes