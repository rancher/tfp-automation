#!/bin/bash

set -euo pipefail

export PREFIX="${PREFIX}"
export PROJECT_ID="${PROJECT_ID}"
export ZONE="${ZONE}"

echo "Cleanup in progress..."

INSTANCES=$(gcloud compute instances list --project="$PROJECT_ID" --filter="name~'^${PREFIX}' AND zone:($ZONE)" --format="value(name)")
if [ -n "$INSTANCES" ]; then
  for INSTANCE in $INSTANCES; do
    echo "Deleting instance: $INSTANCE in zone: $ZONE"
    gcloud compute instances delete "$INSTANCE" --project="$PROJECT_ID" --zone="$ZONE" --quiet > /dev/null 2>&1
  done
else
  echo "No instances found with prefix '$PREFIX' in zone $ZONE for project $PROJECT_ID."
fi

# If prefix does not start with "auto", then we are cleaning up the infrastructure.
if [[ "$PREFIX" != auto* ]]; then
  FIREWALLS=$(gcloud compute firewall-rules list --project="$PROJECT_ID" --filter="name~'^${PREFIX}'" --format="value(name)")
  if [ -n "$FIREWALLS" ]; then
    for firewall in $FIREWALLS; do
      echo "Deleting firewall rule: $firewall"
      gcloud compute firewall-rules delete "$firewall" --project="$PROJECT_ID" --quiet > /dev/null 2>&1
    done
  else
    echo "No firewall rules found with prefix '$PREFIX'."
  fi

  FWD_RULES=$(gcloud compute forwarding-rules list --project="$PROJECT_ID" --filter="name~'^${PREFIX}'" --format="value(name,region)")
  if [ -n "$FWD_RULES" ]; then
    while read -r RULE REGION; do
      echo "Deleting forwarding rule: $RULE in region: $REGION"
      gcloud compute forwarding-rules delete "$RULE" --project="$PROJECT_ID" --region="$REGION" --quiet > /dev/null 2>&1
    done <<< "$FWD_RULES"
  else
    echo "No forwarding rules found with prefix '$PREFIX'."
  fi

  BACKENDS=$(gcloud compute backend-services list --project="$PROJECT_ID" --filter="name~'^${PREFIX}'" --format="value(name,region)")
  if [ -n "$BACKENDS" ]; then
    while read -r BACKEND REGION; do
      if [ -n "$REGION" ]; then
        echo "Deleting regional backend service: $BACKEND in region: $REGION"
        gcloud compute backend-services delete "$BACKEND" --project="$PROJECT_ID" --region="$REGION" --quiet > /dev/null 2>&1
      else
        echo "Deleting global backend service: $BACKEND"
        gcloud compute backend-services delete "$BACKEND" --project="$PROJECT_ID" --global --quiet > /dev/null 2>&1
      fi
    done <<< "$BACKENDS"
  else
    echo "No backend services found with prefix '$PREFIX'."
  fi

  HEALTH_CHECKS=$(gcloud compute health-checks list --project="$PROJECT_ID" --filter="name~'^${PREFIX}'" --format="value(name)")
  if [ -n "$HEALTH_CHECKS" ]; then
    for health_check in $HEALTH_CHECKS; do
      echo "Deleting health check: $health_check"
      gcloud compute health-checks delete "$health_check" --project="$PROJECT_ID" --quiet > /dev/null 2>&1
    done
  else
    echo "No health checks found with prefix '$PREFIX'."
  fi

  INSTANCE_GROUPS=$(gcloud compute instance-groups list --project="$PROJECT_ID" --filter="name~'^${PREFIX}' AND zone:($ZONE)" --format="value(name)")
  if [ -n "$INSTANCE_GROUPS" ]; then
    for instance_group in $INSTANCE_GROUPS; do
      echo "Deleting instance group: $instance_group in zone: $ZONE"
      gcloud compute instance-groups unmanaged delete "$instance_group" --project="$PROJECT_ID" --zone="$ZONE" --quiet > /dev/null 2>&1
    done
  else
    echo "No instance groups found with prefix '$PREFIX' in zone $ZONE."
  fi

  ADDRESSES=$(gcloud compute addresses list --project="$PROJECT_ID" --filter="name~'^${PREFIX}'" --format="value(name,region)")
  if [ -n "$ADDRESSES" ]; then
    while read -r ADDRESS REGION; do
      if [ -n "$REGION" ]; then
        echo "Deleting regional address: $ADDRESS in region: $REGION"
        gcloud compute addresses delete "$ADDRESS" --project="$PROJECT_ID" --region="$REGION" --quiet > /dev/null 2>&1
      else
        echo "Deleting global address: $ADDRESS"
        gcloud compute addresses delete "$ADDRESS" --project="$PROJECT_ID" --global --quiet > /dev/null 2>&1
      fi
    done <<< "$ADDRESSES"
  else
    echo "No addresses found with prefix '$PREFIX'."
  fi
fi

echo "Cleanup completed!"