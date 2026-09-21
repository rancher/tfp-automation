#!/bin/bash

set -euo pipefail

PREFIX="${PREFIX}"
AWS_REGION="${AWS_REGION}"

export AWS_REGION

echo "Cleanup in progress..."

INSTANCE_IDS=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=${PREFIX}*" \
  --query "Reservations[].Instances[?State.Name!='terminated'].InstanceId" \
  --output text)

if [ -n "$INSTANCE_IDS" ]; then
  echo "Deleting EC2 instances..."
  aws ec2 terminate-instances --instance-ids $INSTANCE_IDS > /dev/null 2>&1
else
  echo "No matching instances found"
fi

# If prefix does not start with "auto", then we are cleaning up the infrastructure.
if [[ "$PREFIX" != auto* ]]; then
  KEYPAIRS=$(aws ec2 describe-key-pairs \
    --query "KeyPairs[?starts_with(KeyName, \`${PREFIX}\`)].KeyName" \
    --output text)

  if [ -n "$KEYPAIRS" ]; then
    echo "Deleting EC2 key pairs..."

    for kp in $KEYPAIRS; do
      aws ec2 delete-key-pair --key-name "$kp" > /dev/null 2>&1
    done
  else
    echo "No matching key pairs found."
  fi

  LBS=$(aws elbv2 describe-load-balancers \
    --query "LoadBalancers[?starts_with(LoadBalancerName, \`${PREFIX}\`)].LoadBalancerArn" \
    --output text || true)

  if [ -n "$LBS" ]; then
    for lb in $LBS; do
      echo "Deleting load balancer..."
      aws elbv2 delete-load-balancer --load-balancer-arn "$lb" > /dev/null 2>&1
    done
  else
    echo "No matching load balancers found."
  fi

  TGS=$(aws elbv2 describe-target-groups \
    --query "TargetGroups[?starts_with(TargetGroupName, \`${PREFIX}\`)].TargetGroupArn" \
    --output text || true)

  if [ -n "$TGS" ]; then
    echo "Deleting target groups..."

    for tg in $TGS; do
      aws elbv2 delete-target-group --target-group-arn "$tg" > /dev/null 2>&1
    done
  else
    echo "No matching target groups found."
  fi

  HOSTED_ZONE_IDS=$(aws route53 list-hosted-zones --query "HostedZones[].Id" --output text)

  for zone_id in $HOSTED_ZONE_IDS; do
    RECS=$(aws route53 list-resource-record-sets --hosted-zone-id "$zone_id" \
      --query "ResourceRecordSets[?starts_with(Name, \`${PREFIX}\`)]" \
      --output json)

    if [ "$RECS" != "[]" ]; then
      echo "Deleting Route53 records..."

      CHANGE_BATCH=$(jq -n --argjson recs "$RECS" '
      {
        Changes: ($recs | map({Action: "DELETE", ResourceRecordSet: .}))
      }')

      printf "%s" "$CHANGE_BATCH" > change-batch.json

      aws route53 change-resource-record-sets \
        --hosted-zone-id "$zone_id" \
        --change-batch file://change-batch.json \
        >/dev/null 2>&1
    else
      echo "No matching Route53 records found in zone..."
    fi
  done
fi

echo "Cleanup completed!"