#!/bin/bash

set -x
set -eu

DEBUG="${DEBUG:-false}"

env | egrep '^(CATTLE_|CONFIG|RANCHER2_|BUILD_|BRANCH|TEST_|JOB_|REPO|TIMEOUT).*\=.+' | sort > .env

if [ "false" != "${DEBUG}" ]; then
    cat .env
fi