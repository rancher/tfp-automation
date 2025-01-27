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
  etcd:
    disableSnapshot: false
    snapshotCron: "0 */5 * * *"
    snapshotRetention: 3
  module: # select from: ec2_rke1_custom, ec2_rke2_custom or ec2_k3s_custom
  hostnamePrefix: ""
  machineConfigName: ""
  cni: ""
  cloudCredentialName: ""
  enableNetworkPolicy: false
  defaultClusterRoleForProjectMembers: user
  privateKeyPath: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    awsKeyName: ""
    ami: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroupNames:
      - ""
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    sshTimeout: "5m"
terratest:
  nodeCount: 3
```

See the below examples on how to run the tests:

### RKE1/RKE2/K3S

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionTestSuite/TestTfpProvision$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpProvisionTestSuite/TestTfpProvisionDynamicInput$"`

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