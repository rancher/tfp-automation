#!/bin/bash

RESOURCE_PREFIX=$1
AWS_REGION=$2
AWS_ACCESS_KEY=$3
AWS_SECRET_KEY=$4

set -e

. /etc/os-release

[[ "${ID}" == "ubuntu" || "${ID}" == "debian" ]] && sudo apt install unzip -y > /dev/null
[[ "${ID}" == "opensuse-leap" || "${ID}" == "sles" ]] && sudo zypper install -y unzip > /dev/null

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" ]]; then
    ARCH="x86_64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="aarch64"
fi

echo "Downloading AWS CLI..."
curl "https://awscli.amazonaws.com/awscli-exe-linux-${ARCH}.zip" -o "awscliv2.zip"
unzip awscliv2.zip > /dev/null
sudo ./aws/install > /dev/null
rm -rf aws awscliv2.zip

echo "Configuring AWS CLI..."
aws configure set aws_access_key_id ${AWS_ACCESS_KEY}
aws configure set aws_secret_access_key ${AWS_SECRET_KEY}
aws configure set region ${AWS_REGION}
aws configure set output json

ARCH=$(uname -m)
if [[ $ARCH == "x86_64" || $ARCH == "amd64" ]]; then
    ARCH="amd64"
elif [[ $ARCH == "arm64" || $ARCH == "aarch64" ]]; then
    ARCH="arm64"
fi

PLATFORM=$(uname -s)_$ARCH

echo "Installing kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
mkdir -p ~/.kube
rm kubectl

echo "Installing Helm"
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod +x get_helm.sh
./get_helm.sh
rm get_helm.sh

echo "Downloading eksctl..."
curl -sLO "https://github.com/eksctl-io/eksctl/releases/latest/download/eksctl_$PLATFORM.tar.gz"
tar -xzf eksctl_$PLATFORM.tar.gz -C /tmp && rm eksctl_$PLATFORM.tar.gz
sudo install -m 0755 /tmp/eksctl /usr/local/bin && rm /tmp/eksctl

echo "Creating EKS cluster..."
eksctl create cluster --name ${RESOURCE_PREFIX} --region ${AWS_REGION} --nodegroup-name ${RESOURCE_PREFIX}-ng \
                                                                       --nodes 3 --nodes-min 3 --nodes-max 3 --managed > /dev/null 2>&1

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "Creating service account and associating OIDC provider..."
eksctl utils associate-iam-oidc-provider --region=${AWS_REGION} --cluster=${RESOURCE_PREFIX} --approve
eksctl create iamserviceaccount --cluster=${RESOURCE_PREFIX} \
                                --namespace=kube-system \
                                --name=aws-load-balancer-controller \
                                --attach-policy-arn=arn:aws:iam::${ACCOUNT_ID}:policy/AWSLoadBalancerControllerIAMPolicy \
                                --override-existing-serviceaccounts \
                                --region ${AWS_REGION} --approve

echo "Installing AWS Load Balancer Controller..."
helm repo add eks https://aws.github.io/eks-charts
helm repo update eks
helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
             -n kube-system --set clusterName=${RESOURCE_PREFIX} --set serviceAccount.create=false \
             --set serviceAccount.name=aws-load-balancer-controller

echo "Waiting for AWS Load Balancer Controller to be ready..."
kubectl -n kube-system rollout status deploy/aws-load-balancer-controller