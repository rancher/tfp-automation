# TFP-Automation KDM Test

In the tfp-automation KDM test, the following workflow is followed:

1. Setup Rancher HA utilizing Terraform resources + specified provider infrastructure
2. Verify KDM Url and versions
3. Provision downstream RKE2 / K3S clusters.
4. Perform post-cluster provisioning checks
5. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Local Qase Reporting](#Local-Qase-Reporting)

## Getting Started
The config is split up into multiple parts. Think of the parts as follows:
- Standalone config for setting up Rancher
- Node driver config for provisioning downstream clusters
- Rancher config

In no particular order, see an example below:

```yaml
#######################
# RANCHER CONFIG
#######################
rancher:
  host: ""                                        # REQUIRED - fill out with the expected Rancher server URL
  adminPassword: ""                               # REQUIRED - this is the same as the bootstrapPassword below, make sure they match
  insecure: true                                  # REQUIRED - leave this as true
#######################
# TERRAFORM CONFIG
#######################
terraform:
  cni: ""                                         # REQUIRED - fill with desired value
  defaultClusterRoleForProjectMembers: "true"     # REQUIRED - leave value as true
  enableNetworkPolicy: false                      # REQUIRED - values are true or false -  can leave as false
  provider: "aws"
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""                              # REQUIRED - fill with desired value
  privateRegistries:                              # This is an optional block. You must already have a private registry stood up
    url: ""
    systemDefaultRegistry: ""                     # RKE2/K3S specific, can be left blank
    insecure: true
    authConfigSecretName: ""                      # RKE2/K3S specific
    mirrorHostname: ""
    mirrorEndpoint: ""
  ##########################################
  # STANDALONE / RANCHER CLUSTER CONFIG
  ##########################################
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsSecurityGroups: [""]
    awsSecurityGroupNames: [""]
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    ipAddressType: "ipv4"
    loadBalancerType: "ipv4"
    targetType: "instance"
  ###################################
  # STANDALONE CONFIG - RANCHER SETUP
  ###################################
  standalone:
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    certType: ""                                  # REQUIRED - "self-signed" or "lets-encrypt"
    chartVersion: ""                              # REQUIRED - fill with desired value (leave out the leading 'v')
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using Rancher Prime or staging registry
    rancherChartVersion: ""                       # REQUIRED - fill with desired value
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    registryPassword: ""                          # REQUIRED
    registryUsername: ""                          # REQUIRED
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Group: ""                                 # REQUIRED - fill with group of the instance created
    rke2User: ""                                  # REQUIRED - fill with username of the instance created
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value (i.e. v1.32.6)
#######################
# TERRATEST CONFIG
#######################
terratest:  
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
```

Before running, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export CLOUD_PROVIDER_VERSION=""
export LETS_ENCRYPT_EMAIL=""                      # OPTIONAL - must provide a valid email address
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/kdm --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestKDMTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/kdm  --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestKDMTestSuite$";/path/to/tfp-automation/reporter`