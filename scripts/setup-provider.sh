#!/bin/bash

# QA has a new process to test Terraform using RCs. We will not publish RCs
# onto the Terraform registry because Hashicorp could potentially block our
# testing by being slow to publish.

# Instead, we will test using a downloaded binary from the RC. This script
# sets up the correct binary using a defined <provider> <version> to test
# updates locally.

# ./setup-provider.sh <provider> <version>

# Example
# ./setup-provider.sh rancher2 v13.0.0-rc.6

set -e 

if [ $# -ne 2 ]; then
  echo "Usage: $0 <provider> <version>"
  exit 1
fi

PROVIDER=$1
VERSION=$2
VERSION_TAG=$(echo $2 | cut -c 2-)

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
  PLATFORM="darwin_arm64"
else
  PLATFORM="linux_amd64"
fi

DIR=~/.terraform.d/plugins/terraform.local/local/${PROVIDER}/${VERSION_TAG}/${PLATFORM}
(umask u=rwx,g=rwx,o=rwx && mkdir -p $DIR)
curl -sfL https://github.com/rancher/terraform-provider-${PROVIDER}/releases/download/${VERSION}/terraform-provider-${PROVIDER}_${VERSION_TAG}_${PLATFORM}.zip -o /tmp/provider.zip
unzip -o /tmp/provider.zip -d ${DIR}

BINARY=$(find ${DIR} -type f -name "terraform-provider-${PROVIDER}_v${VERSION_TAG}")

if [ -f "$BINARY" ]; then
  mv "$BINARY" "${DIR}/terraform-provider-${PROVIDER}"
  chmod +x "${DIR}/terraform-provider-${PROVIDER}"
else
  echo "ERROR: Provider binary not found after unzip!"
  exit 1
fi

cat <<EOF
Terraform provider ${PROVIDER} ${VERSION} is ready to test!
Please update the required_providers block in your Terraform config file

terraform {
  required_providers {
    rancher2 = {
      source = "terraform.local/local/${PROVIDER}"
      version = "${VERSION_TAG}"
    }
  }
}
EOF