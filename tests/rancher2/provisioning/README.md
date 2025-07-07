# Provisioning

In the provisioning tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Provisioning Clusters](#Provisioning-Clusters)
3. [Local Qase Reporting](#Local-Qase-Reporting)

## Getting Started
In your config file, set the following:
```yaml
rancher:
  host: "rancher_server_address"
  adminToken: "rancher_admin_token"
  insecure: true
  cleanup: true
```

To see what goes into the `terraform` block in addition to the `rancher`, please refer to the tfp-automation [README](../../README.md).

## Provisioning Clusters
The provisioning tests have static test cases as well as dynamicInput tests you can specify. In order to run the dynamicInput tests, you will need to define the `terratest` block in your config file. See an example below:

```yaml
terratest:
  kubernetesVersion: ""
  psact: "" # Optional, can be left out or can have values `rancher-privileged` or `rancher-restricted`
  ```

For provisioning with custom clusters, reference the example config block below:

```yaml
terraform:
  cni: ""
  enableNetworkPolicy: false
  defaultClusterRoleForProjectMembers: "user"
  module:                       # ec2_rke1_custom, ec2_rke2_custom, ec2_k3s_custom, vsphere_rke1_custom, vsphere_rke2_custom, vsphere_k3s_custom
  privateKeyPath: ""
  provider: ""                  # aws or vsphere
  windowsPrivateKeyPath: ""
  
  # Set if provider: aws
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    awsKeyName: ""
    ami: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    sshTimeout: "5m"
    windowsAMI: ""
    windowsAWSUser: ""
    windowsInstanceType: ""
    windowsKeyName: ""
  
  # Set if provider: vsphere
  vsphereCredentials:
    password: ""
    username: ""
    vcenter: ""
  vsphereConfig:  
    cloneFrom: ""
    cpuCount: ""
    datacenter: ""
    datastore: ""
    datastoreCluster: ""
    diskSize: ""
    guestID: ""
    folder: ""
    hostSystem: ""
    memorySize: ""
    standaloneNetwork: ""
    vsphereUser: ""
terratest:
  nodeCount: 3
  windowsNodeCount: 1
```

For running the imported clusters, reference the example config block below:

```yaml
rancher:
  host: ""
  adminToken: ""
  cleanup: true

terraform:
  cni: ""
  defaultClusterRoleForProjectMembers: "true"
  enableNetworkPolicy: false
  module:                          # ec2_rke1_import, ec2_rke2_import, ec2_k3s_import, vsphere_rke1_import, vsphere_rke2_import, vsphere_k3s_import
  privateKeyPath: ""
  provider: ""                     # aws or vsphere
  windowsPrivateKeyPath: ""

  # Set if provider: aws
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""

  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    windowsAMI: ""
    windowsAWSUser: ""
    windowsAWSPassword: ""
    windowsInstanceType: ""
    windowsKeyName: ""

  # Set if provider: vsphere
  vsphereCredentials:
    password: ""
    username: ""
    vcenter: ""

  vsphereConfig:  
    cloneFrom: ""
    cpuCount: ""
    datacenter: ""
    datastore: ""
    datastoreCluster: ""
    diskSize: ""
    guestID: ""
    folder: ""
    hostSystem: ""
    memorySize: ""
    standaloneNetwork: ""
    vsphereUser: ""

  standalone:
    k3sVersion: ""                      # Ensure k3s1 suffix is appended (i.e. v1.xx.x+k3s1)
    osGroup: ""
    osUser: ""
    rancherHostname: ""
    rke2Version: ""                     # Ensure rke2r1 suffix is appended (i.e. v1.xx.x+rke2r1)

terratest:
  tfLogging: true
  nodeCount: 3
  windowsNodeCount: 1
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
```

In addition, when running locally, you will need to ensure that you have `export RKE_PROVIDER_VERSION=x.x.x` defined for the RKE1 portion of the test. You also must ensure that you are not using the highest available K8s version as this test will perform an upgrade of the imported cluster.

See the below examples on how to run the tests:

### RKE1/RKE2/K3S

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionTestSuite/TestTfpProvision$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionTestSuite/TestTfpProvisionDynamicInput$"`

### Custom
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionCustomTestSuite/TestTfpProvisionCustom$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionCustomTestSuite/TestTfpProvisionCustomDynamicInput$"`

### Imported

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpUpgradeImportedClusterTestSuite/TestTfpUpgradeImportedCluster$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpUpgradeImportedClusterTestSuite/TestTfpUpgradeImportedClusterDynamicInput$"`

### Hosted

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionHostedTestSuite/TestTfpProvisionHosted$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionTestSuite/TestTfpProvision$";/path/to/tfp-automation/reporter`