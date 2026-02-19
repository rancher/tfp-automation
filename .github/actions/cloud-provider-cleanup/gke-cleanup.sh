#!/bin/bash

set -euo pipefail

export PREFIX="${PREFIX}"
export PROJECT_ID="${PROJECT_ID}"
export ZONE="${ZONE}"

echo "Cleanup in progress..."

CLUSTERS=$(gcloud container clusters list --project="${PROJECT_ID}" --filter="name~'^${PREFIX}' AND zone:($ZONE)" --format="value(name)")
if [ -n "$CLUSTERS" ]; then
  for CLUSTER in $CLUSTERS; do
    echo "Deleting cluster: ${CLUSTER} in zone: ${ZONE}"
    gcloud container clusters delete "${CLUSTER}" --zone="${ZONE}" --quiet > /dev/null 2>&1 &
  done
else
  echo "No matching clusters found."
fi

INSTANCES=$(gcloud compute instances list --project="${PROJECT_ID}" --filter="name~'^${PREFIX}' AND zone:($ZONE)" --format="value(name)")
if [ -n "$INSTANCES" ]; then
  for INSTANCE in $INSTANCES; do
    echo "Deleting instance: ${INSTANCE} in zone: ${ZONE}"
    gcloud compute instances delete "${INSTANCE}" --project="${PROJECT_ID}" --zone="${ZONE}" --quiet > /dev/null 2>&1 &
  done
else
  echo "No matching instances found."
fi

echo "Cleanup completed!"