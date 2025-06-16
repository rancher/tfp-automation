# PSACT

In the psact tests, the following workflow is followed:

1. Provision a downstream cluster with rancher-privileged
2. Perform post-cluster provisioning checks
3. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)
4. Provision a downstream cluster with rancher-restricted
5. Perform post-cluster provisioning checks
6. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)
7. Provision a downstream cluster with rancher-baseline
8. Perform post-cluster provisioning checks
9. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

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
The psact tests are static test cases, so the config to be provided will exactly match that of the provisioning test config. See an example below:

```yaml
terraform:
  cloudCredentialName: "tfp-creds"
  defaultClusterRoleForProjectMembers: "true"
  enableNetworkPolicy: false
  hostnamePrefix: "tfp-automation"
  machineConfigName: "tfp-automation"
  module: "linode_k3s"
  linodeConfig:
    linodeToken: ""
    linodeImage: "linode/ubuntu22.04"
    region: "us-east"
    linodeRootPass: "<placeholder>"
terratest:
  kubernetesVersion: ""
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
  ```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/psact --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpPSACTTestSuite/TestTfpPSACT$"`

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/psact --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpPSACTTestSuite/TestTfpPSACT$";/path/to/tfp-automation/reporter`