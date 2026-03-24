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
  adminPassword: "rancher_admin_password"
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
  downstreamClusterProvider: ""       # REQUIRED - can be aws, azure, linode, vsphere
  localAuthEndpoint: false      # OPTIONAL - false by default
  privateKeyPath: ""
  provider: ""                  # aws or vsphere
  windowsPrivateKeyPath: ""
  dataDirectories:              # OPTIONAL - configure for custom data directory test only
    systemAgentPath: ""
    provisioningPath: ""
    k8sDistroPath: ""
  
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
    rancherSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    sshTimeout: "5m"
    windowsAMI2019: ""
    windowsAMI2022: ""
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
    firmware: ""               # OPTIONAL - set to "efi" for UEFI templates
    guestID: ""
    folder: ""
    hostSystem: ""
    memorySize: ""
    pool: ""                   # OPTIONAL - existing resource pool name/path
    standaloneNetwork: ""
    vsphereUser: ""
terratest:
  etcdCount: 3
  controlPlaneCount: 2
  workerCount: 3
  windowsNodeCount: 1
```

For running the imported clusters, reference the example config block below:

```yaml
rancher:
  host: ""
  adminPassword: ""
  cleanup: true
  insecure: true

terraform:
  cni: ""
  defaultClusterRoleForProjectMembers: "true"
  downstreamClusterProvider: ""       # REQUIRED - can be aws, azure, linode, vsphere
  enableNetworkPolicy: false
  mixedArchitecture:                              # OPTIONAL - set to true if you want mixed architecture
  privateKeyPath: ""
  provider: ""                     # aws or vsphere
  windowsPrivateKeyPath: ""
  dataDirectories:                 # OPTIONAL - configure for custom data directory test only
    systemAgentPath: ""
    provisioningPath: ""
    k8sDistroPath: ""

  # Set if provider: aws
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""

  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    armAMI: ""                                    # OPTIONAL - only set if mixedArchitecture is set to true
    armInstanceType: ""                           # OPTIONAL - only set if mixedArchitecture is set to true
    region: ""
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    rancherSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    windowsAMI2019: ""
    windowsAMI2022: ""
    windowsAWSUser: ""
    windows2019Password: ""
    windows2022Password: ""
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
    firmware: ""               # OPTIONAL - set to "efi" for UEFI templates
    guestID: ""
    folder: ""
    hostSystem: ""
    memorySize: ""
    pool: ""                   # OPTIONAL - existing resource pool name/path
    standaloneNetwork: ""
    vsphereUser: ""

  standalone:
    k3sVersion: ""                      # Ensure k3s1 suffix is appended (i.e. v1.xx.x+k3s1)
    osGroup: ""
    osUser: ""
    rancherHostname: ""
    rke2Version: ""                     # Ensure rke2r1 suffix is appended (i.e. v1.xx.x+rke2r1)

terratest:
  etcdCount: 3
  controlPlaneCount: 2
  workerCount: 3
  windowsNodeCount: 1
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
```

If you would like a private registry associated to your downstream cluster, enter in the optional parameters underneath the `terraform` block:

```yaml
privateRegistries:                          # This is an optional block. You must already have a private registry stood up
  url: ""
  systemDefaultRegistry: ""                 # OPTIONAL
  username: ""
  password: ""
  insecure: true
  authConfigSecretName: ""                  # OPTIONAL
  mirrorHostname: ""
  mirrorEndpoint: ""
  mirrorRewrite: ""
```

See the below examples on how to run the tests:

### Node Driver
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpProvisionTestSuite/TestTfpProvision$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=dynamic -v -run "TestTfpProvisionTestSuite/TestTfpProvisionDynamicInput$"`

### Custom
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpProvisionCustomTestSuite/TestTfpProvisionCustom$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=dynamic -v -run "TestTfpProvisionCustomTestSuite/TestTfpProvisionCustomDynamicInput$"`

### Imported
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpUpgradeImportedClusterTestSuite/TestTfpUpgradeImportedCluster$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=dynamic -v -run "TestTfpUpgradeImportedClusterTestSuite/TestTfpUpgradeImportedClusterDynamicInput$"`

### Hosted
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpProvisionHostedTestSuite/TestTfpProvisionHosted$"`

### ACE
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestProvisionACETestSuite/TestTfpProvisionACE$"`

### Data Directories
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestProvisionDataDirectoryTestSuite/TestTfpProvisionDataDirectory$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpProvisionTestSuite/TestTfpProvision$";/path/to/tfp-automation/reporter`