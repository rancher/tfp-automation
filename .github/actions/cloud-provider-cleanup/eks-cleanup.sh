#!/bin/bash

set -euo pipefail

export PREFIX="${PREFIX}"
export AWS_REGION="${AWS_REGION}"

echo "Cleanup in progress..."

ZONE=$(aws route53 list-hosted-zones --query "HostedZones[0].Name" --output text | sed 's/\.$//')
HOSTED_ZONE_ID=$(aws route53 list-hosted-zones-by-name --query "HostedZones[0].Id" --output text | sed 's|/hostedzone/||')
ROUTE53_RECORD=$(aws route53 list-resource-record-sets --hosted-zone-id "${HOSTED_ZONE_ID}" --query "ResourceRecordSets[?starts_with(Name, \`${PREFIX}\`)].Name" --output text | sed 's/\.$//')

if [ -n "$ROUTE53_RECORD" ]; then
  echo "Fetching current Route53 record values for deletion..."
  CURRENT_RECORD_JSON=$(aws route53 list-resource-record-sets --hosted-zone-id "${HOSTED_ZONE_ID}" --query "ResourceRecordSets[?Name=='${ROUTE53_RECORD}.'] | [?Type=='CNAME']" --output json)
  CURRENT_VALUE=$(echo "$CURRENT_RECORD_JSON" | jq -r '.[0].ResourceRecords[0].Value')

  if [ -n "$CURRENT_VALUE" ] && [ "$CURRENT_VALUE" != "null" ]; then
    echo "Deleting Route53 record with current value: $CURRENT_VALUE"
    aws route53 change-resource-record-sets \
      --hosted-zone-id "${HOSTED_ZONE_ID}" \
      --change-batch "{\"Changes\":[{\"Action\":\"DELETE\",\"ResourceRecordSet\":{\"Name\":\"${ROUTE53_RECORD}.\",\"Type\":\"CNAME\",\"TTL\":300,\"ResourceRecords\":[{\"Value\":\"${CURRENT_VALUE}\"}]}}]}" \
      > /dev/null 2>&1
  else
    echo "No current value found for Route53 record, skipping deletion."
  fi
else
  echo "No matching Route53 records found."
fi

CLUSTER_NAMES=$(aws eks list-clusters --query "clusters[?starts_with(@, \`${PREFIX}\`)]" --output text)

if [ -n "$CLUSTER_NAMES" ]; then
  for cluster in $CLUSTER_NAMES; do
    echo "Deleting cluster $cluster..."
    nohup eksctl delete cluster --name "$cluster" --region "$AWS_REGION" > /dev/null 2>&1 &
  done
else
  echo "No matching clusters found."
fi

VPC_IDS=$(aws ec2 describe-vpcs --filters "Name=tag:Name,Values=eksctl-${PREFIX}*" --query "Vpcs[?State=='available'].VpcId" --output text)
SECURITY_GROUPS=$(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=${VPC_IDS}" --query "SecurityGroups[].GroupId" --output text)

if [ -n "$SECURITY_GROUPS" ]; then
  for sg in $SECURITY_GROUPS; do
    echo "Deleting security group $sg..."
    aws ec2 delete-security-group --group-id "$sg" > /dev/null 2>&1
  done
else
  echo "No matching security groups found."
fi

if [ -n "$VPC_IDS" ]; then
  for vpc in $VPC_IDS; do
    echo "Deleting VPC $vpc..."
    aws ec2 delete-vpc --vpc-id "$vpc" > /dev/null 2>&1
  done
else
  echo "No matching VPCs found."
fi

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

echo "Cleanup completed!"