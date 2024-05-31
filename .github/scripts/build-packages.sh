#!/bin/bash
set -e

oldPWD="$(pwd)"

dirs=("./config" "./defaults" "./framework" "./modules" "./pipeline" "./scripts" "./tests")

for dir in "${dirs[@]}"; do
    echo "Building $dir"
    cd "$dir"
    go build ./...
    cd "$oldPWD"
done