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
  insecure: true                                  # REQUIRED - leave this as true
#######################
# TERRAFORM CONFIG
#######################
terraform:
  cni: ""                                         # REQUIRED - fill with desired value
  defaultClusterRoleForProjectMembers: "true"     # REQUIRED - leave value as true
  downstreamClusterProvider: "aws"
  enableNetworkPolicy: false                      # REQUIRED - values are true or false -  can leave as false
  mixedArchitecture:                              # OPTIONAL - set to true if you want mixed architecture
  provider: "aws"
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""                              # REQUIRED - fill with desired value
  windowsPrivateKeyPath: ""                       # REQUIRED - specify Windows private key that will be used to access created instances
  privateRegistries:                              # This is an optional block. You must already have a private registry stood up
    engineInsecureRegistry: ""                    # RKE1 specific
    url: ""
    systemDefaultRegistry: ""                     # RKE2/K3S specific, can be left blank
    username: ""                                  # RKE1 specific
    password: ""                                  # RKE1 specific
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
    armAMI: ""                                    # OPTIONAL - only set if mixedArchitecture is set to true
    armInstanceType: ""                           # OPTIONAL - only set if mixedArchitecture is set to true
    awsSubnetID: ""
    rancherSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsSecurityGroups: [""]
    awsSecurityGroupNames: [""]
    clusterCIDR: ""             # OPTIONAL - Use for IPv6/dual-stack
    serviceCIDR: ""             # OPTIONAL - Use for IPv6/dual-stack
    enablePrimaryIPv6: true     # OPTIONAL - Use for IPv6/dual-stack (true or false)
    httpProtocolIpv6: ""        # OPTIONAL - Use for IPv6/dual-stack (enabled or disabled)
    ipv6AddressOnly: true       # OPTIONAL - Use for IPv6/dual-stack (true or false)
    ipv6AddressCount: "1"       # OPTIONAL - Use for IPv6/dual-stack
    networking:
      stackPreference: ""       # OPTIONAL - USE for IPv6/dual-stack (ipv6 or dual)
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    ipAddressType: "ipv4"
    loadBalancerType: "ipv4"
    targetType: "instance"
    windowsAMI2019: ""
    windowsAMI2022: ""
    windowsAWSUser: ""
    windows2019Password: ""
    windows2022Password: ""
    windowsInstanceType: ""
    windowsKeyName: ""
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
    upgradedRancherAgentImage: ""                 # OPTIONAL - fill out if you are performing an upgrade
    upgradedRancherChartRepository: ""            # OPTIONAL - fill out if you are performing an upgrade
    upgradedRancherChartVersion: ""               # OPTIONAL - fill out if you are performing an upgrade
    upgradedRancherImage: ""                      # OPTIONAL - fill out if you are performing an upgrade
    upgradedRancherRepo: ""                       # OPTIONAL - fill out if you are performing an upgrade
    upgradedRancherTagVersion: ""                 # OPTIONAL - fill out if you are performing an upgrade
    featureFlags:
      turtles: ""                                 # REQUIRED - "true", "false", "toggledOn", or "toggledOff"
      upgradedTurtles: ""                         # REQUIRED - "true", "false", "toggledOn", or "toggledOff"
#######################
# TERRATEST CONFIG
#######################
terratest:  
  etcdCount: 3
  controlPlaneCount: 2
  workerCount: 3
  windowsNodeCount: 1
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

### RKE2 / K3s
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityProvisioningTestSuite$"`

### IPv6
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityIPv6ProvisioningTestSuite$"`

### Dual-stack
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityDualStackProvisioningTestSuite$"`

### AKS
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityAKSProvisioningTestSuite$"`

### Upgrade
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityUpgradeRancherTestSuite$"`

### Upgrade IPv6
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityIPv6UpgradeRancherTestSuite$"`

### Upgrade Dual-stack
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityDualStackUpgradeRancherTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/sanity  --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestTfpSanityProvisioningTestSuite$";/path/to/tfp-automation/reporter`