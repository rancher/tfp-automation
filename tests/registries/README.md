# Private Registry Tests

In the tfp-automation private registries test, the following workflow is followed:

1. Setup Rancher HA utilizing Terraform resources + specified provider infrastructure. A global registry is set as the system default registry while an authenticated and non-authenticated registry are created.
2. Provision downstream RKE1 / RKE2 / K3S clusters - done using the global registry, authenticated registry and non-authenticated registry.
3. Perform post-cluster provisioning checks
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Local Qase Reporting](#Local-Qase-Reporting)

## Getting Started
The config is split up into multiple parts. Think of the parts as follows:
- Standalone config for setting up Rancher
- Standalone config for setting up private registry
- Custom cluster config for provisioning downstream clusters
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
  privateRegistries:
    authConfigSecretName: ""                      # REQUIRED (authenticated registry only) - specify the name of the secret you wanted created
    insecure: true
    username: ""                                  # REQUIRED (authenticated registry only) - username of the private registry
    password: ""                                  # REQUIRED (authenticated registry only) - password of the private registry
  resourcePrefix: ""                              # REQUIRED - fill with desired value
  ########################
  # INFRASTRUCTURE SETUP
  ########################
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    awsSecurityGroups: [""]
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    region: ""
    awsUser: ""
    registryRootSize: 500
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
    osGroup: ""                                   # REQUIRED - fill with group of the instance created
    osUser: ""                                    # REQUIRED - fill with username of the instance created
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using Rancher Prime or staging registry
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value (i.e. v1.30.6+rke2r1)
  ####################################
  # STANDALONE CONFIG - REGISTRY SETUP
  ####################################
  standaloneRegistry:
    assetsPath: ""                                # REQUIRED - ensure that you end with a trailing `/`
    registryName: ""                              # REQUIRED (authenticated registry only)
    registryPassword: ""                          # REQUIRED (authenticated registry only)
    registryUsername: ""                          # REQUIRED (authenticated registry only)
    ecrPassword: ""                               # REQUIRED (ecr registry only)
    ecrUsername: ""                               # REQUIRED (ecr registry only)
    ecrURI: ""                                    # REQUIRED (ecr registry only)
    ecrAMI: ""                                    # REQUIRED (ecr registry only) - with Amazon ECR Credential Helper
```

Before running, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export CLOUD_PROVIDER_VERSION=""`
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/registries --junitfile results.xml --jsonfile results.json -- -timeout=8h -v -run "TestTfpRegistriesTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/airgap  --junitfile results.xml --jsonfile results.json -- -timeout=8h -v -run TestTfpAirgapProvisioningTestSuite$";/path/to/tfp-automation/reporter`