# RKE Terraform Provider

In the RKE terraform provider test, the following workflow is followed:

1. Setup 3 AWS instances.
2. Create an RKE 3-node cluster, each with the following roles: controlplane, etcd and worker
3. Validate nodes in the cluster are active

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Local Qase Reporting](#Local-Qase-Reporting)

## Getting Started
The config is split up into multiple parts. Think of the parts as follows:
- Rancher config
- Standalone config for setting up the RKE cluster

See an example below:

```yaml
#######################
# RANCHER CONFIG
#######################
rancher:
  cleanup: true                                     # REQUIRED - leave this as true
#######################
# TERRAFORM CONFIG
#######################
terraform:
  privateKeyPath: ""                                # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""                                # REQUIRED - fill with desired value
  ##########################################
  # STANDALONE CONFIG - INFRASTRUCTURE SETUP
  ##########################################
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: "us-east-2"
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsSecurityGroups: [""]
    awsRootSize: 100
    awsRoute53Zone: ""
    region: ""
    prefix: ""
    awsUser: ""
    sshConnectionType: "ssh"
    standaloneSecurityGroupNames: [""]
    timeout: "5m"
  standalone:
    osUser: ""
```

Before running, be sure to run the following commands:

```yaml
export RKE_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export CLOUD_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rke --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestRKEProviderTestSuite/TestCreateRKECluster$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rke --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestRKEProviderTestSuite/TestCreateRKECluster$";/path/to/tfp-automation/reporter`