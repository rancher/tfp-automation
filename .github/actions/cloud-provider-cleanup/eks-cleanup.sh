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
    eksctl utils write-kubeconfig --cluster "$cluster" --region "$AWS_REGION" > /dev/null
    eksctl delete cluster --name "$cluster" --region "$AWS_REGION" --force --disable-nodegroup-eviction > /dev/null
  done
else
  echo "No matching clusters found."
fi

INSTANCE_IDS=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=${PREFIX}*" \
  --query "Reservations[].Instances[?State.Name!='terminated'].InstanceId" \
  --output text)

if [ -n "$INSTANCE_IDS" ]; then
  echo "Deleting EC2 instances..."
  aws ec2 terminate-instances --instance-ids $INSTANCE_IDS > /dev/null
else
  echo "No matching instances found"
fi

VPC_IDS=$(aws ec2 describe-vpcs --filters "Name=tag:Name,Values=eksctl-${PREFIX}*" --query "Vpcs[?State=='available'].VpcId" --output text)

if [ -n "$VPC_IDS" ]; then
  for vpc in $VPC_IDS; do
    ALB_ARNs=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?VpcId=='$vpc'].LoadBalancerArn" --output text)

    if [ -n "$ALB_ARNs" ]; then
      for alb in $ALB_ARNs; do
        echo "Deleting load balancer in VPC..."
        aws elbv2 delete-load-balancer --load-balancer-arn "$alb" > /dev/null
      done
    else
      echo "No load balancers found."
    fi

    TG_ARNs=$(aws elbv2 describe-target-groups --query "TargetGroups[?VpcId=='$vpc'].TargetGroupArn" --output text)

    if [ -n "$TG_ARNs" ]; then
      for tg in $TG_ARNs; do
        echo "Deleting target group in VPC..."
        aws elbv2 delete-target-group --target-group-arn "$tg" > /dev/null
      done
    else
      echo "No target groups found."
    fi
  done
fi

echo "Cleanup completed!"