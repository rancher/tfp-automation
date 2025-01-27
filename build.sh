#!/bin/bash

set -x
set -eu

DEBUG="${DEBUG:-false}"
TERRAFORM_VERSION="${TERRAFORM_VERSION:-}"
EXTERNAL_ENCODED_VPN="${EXTERNAL_ENCODED_VPN:-}"
VPN_ENCODED_LOGIN="${VPN_ENCODED_LOGIN:-}"
RKE_PROVIDER_VERSION="${RKE_PROVIDER_VERSION:-}"
RANCHER2_PROVIDER_VERSION="${RANCHER2_PROVIDER_VERSION:-}"
LOCALS_PROVIDER_VERSION="${LOCALS_PROVIDER_VERSION:-}"
AWS_PROVIDER_VERSION="${AWS_PROVIDER_VERSION:-}"
RANCHER2_KEY_PATH="${RANCHER2_KEY_PATH:-}"
RKE_KEY_PATH="${RKE_KEY_PATH:-}"
SANITY_KEY_PATH="${SANITY_KEY_PATH:-}"
AIRGAP_KEY_PATH="${AIRGAP_KEY_PATH:-}"
REGISTRY_KEY_PATH="${REGISTRY_KEY_PATH:-}"

TRIM_JOB_NAME=$(basename "$JOB_NAME")

if [ "false" != "${DEBUG}" ]; then
    echo "Environment:"
    env | sort
fi

count=0
while [[ 3 -gt $count ]]; do
    docker build . -f Dockerfile --build-arg CONFIG_FILE=config.yml --build-arg PEM_FILE=key.pem \
                                                                    --build-arg TERRAFORM_VERSION="$TERRAFORM_VERSION" \
                                                                    --build-arg RKE_PROVIDER_VERSION="$RKE_PROVIDER_VERSION" \
                                                                    --build-arg RANCHER2_PROVIDER_VERSION="$RANCHER2_PROVIDER_VERSION" \
                                                                    --build-arg LOCALS_PROVIDER_VERSION="$LOCALS_PROVIDER_VERSION" \
                                                                    --build-arg AWS_PROVIDER_VERSION="$AWS_PROVIDER_VERSION" \
                                                                    --build-arg RANCHER2_KEY_PATH="$RANCHER2_KEY_PATH" \
                                                                    --build-arg RKE_KEY_PATH="$RKE_KEY_PATH" \
                                                                    --build-arg SANITY_KEY_PATH="$SANITY_KEY_PATH" \
                                                                    --build-arg AIRGAP_KEY_PATH="$AIRGAP_KEY_PATH" \
                                                                    --build-arg REGISTRY_KEY_PATH="$REGISTRY_KEY_PATH" \
                                                                    --build-arg EXTERNAL_ENCODED_VPN="$EXTERNAL_ENCODED_VPN" \
                                                                    --build-arg VPN_ENCODED_LOGIN="$VPN_ENCODED_LOGIN" \
                                                                    -t tfp-automation-validation-"${TRIM_JOB_NAME}""${BUILD_NUMBER}"

    if [[ $? -eq 0 ]]; then break; fi
    count=$(($count + 1))
    echo "Repeating failed Docker build ${count} of 3..."
done