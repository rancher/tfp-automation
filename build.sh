#!/bin/bash

set -x
set -eu

DEBUG="${DEBUG:-false}"
RANCHER2_PROVIDER_VERSION="${RANCHER2_PROVIDER_VERSION}"

TRIM_JOB_NAME=$(basename "$JOB_NAME")

if [ "false" != "${DEBUG}" ]; then
    echo "Environment:"
    env | sort
fi

count=0
while [[ 3 -gt $count ]]; do
    docker build . -f Dockerfile --build-arg CONFIG_FILE=config.yml --build-arg RANCHER2_PROVIDER_VERSION="$RANCHER2_PROVIDER_VERSION" -t rancher-validation-"${TRIM_JOB_NAME}""${BUILD_NUMBER}"   

    if [[ $? -eq 0 ]]; then break; fi
    count=$(($count + 1))
    echo "Repeating failed Docker build ${count} of 3..."
done