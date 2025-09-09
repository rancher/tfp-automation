#!/bin/bash
set -e

OS=$(uname | tr '[:upper:]' '[:lower:]')
cd $(dirname $0)/../../

if [[ -z "${QASE_TEST_RUN_ID}" ]]; then
  echo "No QASE test run ID provided"
else
  echo "Building QASE reporter binary"
  env GOOS=${OS} GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o reporter ./pipeline/qase/reporter
fi