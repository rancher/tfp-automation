#!/bin/bash
set -e

OS=$(uname | tr '[:upper:]' '[:lower:]')
cd $(dirname $0)/../../

if [[ -z ${QASE_TEST_RUN_ID} ]]; then
  echo "no test run ID is provided"
else
  echo "building qase reporter bin"
  env GOOS=${OS} GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o reporter ./pipeline/qase/reporter
fi