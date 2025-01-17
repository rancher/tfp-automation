# TFP-Automation Provisioning Sanity

In the tfp-automation provisioning sanity test, the following workflow is followed:

1. Setup Rancher HA utilizing Terraform resources + specified provider infrastructure
2. Provision downstream RKE1 / RKE2 / K3S clusters.
3. Perform post-cluster provisioning checks
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

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
  adminToken: ""                                  # REQUIRED - leave this field empty as shown
  insecure: true                                  # REQUIRED - leave this as true
#######################
# TERRAFORM CONFIG
#######################
terraform:
  cloudCredentialName: ""                         # REQUIRED - fill with desired value
  defaultClusterRoleForProjectMembers: "true"     # REQUIRED - leave value as true
  enableNetworkPolicy: false                      # REQUIRED - values are true or false -  can leave as false
  hostnamePrefix: ""                              # REQUIRED - fill with desired value
  machineConfigName: ""                           # REQUIRED - fill with desired value
  module: ""                                      # REQUIRED - leave this field empty as shown
  networkPlugin: ""                               # REQUIRED - fill with desired value
  nodeTemplateName: ""                            # REQUIRED - fill with desired value
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  ###########################
  # DOWNSTREAM CLUSTER CONFIG
  ###########################
  linodeCredentials:
    linodeToken: ""
  linodeConfig:
    linodeImage: ""
    region: ""
    linodeRootPass: ""
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
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    region: ""
    prefix: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
  ###################################
  # STANDALONE CONFIG - RANCHER SETUP
  ###################################
  standalone:
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    rancherChartVersion: ""                       # REQUIRED - fill with desired value
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherRepo: ""                               # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    rke2Group: ""                                 # REQUIRED - fill with group of the instance created
    type: ""                                      # REQUIRED - fill with desired value
    rke2User: ""                                  # REQUIRED - fill with username of the instance created
    stagingRancherAgentImage: ""                  # OPTIONAL - fill out only if you are using staging registry
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value you wish the local cluster to be
```

Before running, be sure to run the following commands:

`export RANCHER2_KEY_PATH="/<path>/<to>/go/src/github.com/rancher/tfp-automation/modules/rancher2"; export SANITY_KEY_PATH="/<path>/<to>/go/src/github.com/rancher/tfp-automation/modules/sanity"; export RANCHER2_PROVIDER_VERSION=""; export CATTLE_TEST_CONFIG=<path/to/yaml>; export LOCALS_PROVIDER_VERSION=""; export AWS_PROVIDER_VERSION=""`

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=120m -v -run "TestTfpSanityTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity  --junitfile results.xml --jsonfile results.json -- -timeout=120m -v -run TestTfpSanityTestSuite$";/path/to/tfp-automation/reporter`