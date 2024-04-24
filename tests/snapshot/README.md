# Snapshot

In the snapshot tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Perform etcd snapshot
4. Perform post etcd snapshot checks
5. Perform etcd restore
6. Perform post etcd restore checks
7. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

NOTE: Only RKE2/K3s clusters are supported in this package - RKE1 clusters are NOT supported. For reference, see this [ticket](https://github.com/rancher/terraform-provider-rancher2/issues/1292). 

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [ETCD Snapshots](#ETCD-Snapshots)

## Getting Started
In your config file, set the following:
```yaml
rancher:
  host: "rancher_server_address"
  adminToken: "rancher_admin_token"
  insecure: true
  cleanup: true
terraform:
  etcd:
    disableSnapshot: false
    snapshotCron: "0 */5 * * *"
    snapshotRetention: 3
    s3:
      bucket: ""
      cloudCredentialName: ""
      endpoint: "s3.us-east-2.amazonaws.com"
      endpointCA: ""
      folder: ""
      region: "us-east-2"
      skipSSLVerify: true
```

To see what goes into the `terraform` block in addition to the `rancher`, please refer to the tfp-automation [README](../../README.md).

## ETCD Snapshots
Similar to the `provisioning` tests, the snapshot tests have static test cases as well as dynamicInput tests you can specify. In order to run the dynamicInput tests, you will need to define the `terratest` block in your config file. See an example below:

```yaml
terratest:
  kubernetesVersion: ""
  snapshotInput:
    upgradeKubernetesVersion: "" # If left blank, the default version in Rancher will be used.
    snapshotRestore: "all" # Options include none, kubernetesVersion, all. Option 'none' means that only the etcd will be restored.
    controlPlaneConcurrencyValue: "15%"
    workerConcurrencyValue: "20%"
  ```

See the below examples on how to run the tests:

### Snapshot restore
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreTestSuite/TestTfpSnapshotRestoreETCDOnly$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreTestSuite/TestTfpSnapshotRestoreETCDOnlyDynamicInput$"`

### Snapshot restore with K8s upgrade
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreK8sUpgradeTestSuite/TestTfpSnapshotRestoreK8sUpgrade$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreK8sUpgradeTestSuite/TestTfpSnapshotRestoreK8sUpgradeDynamicInput$"`

### Sanpshot restore with upgrade strategy
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreUpgradeStrategyTestSuite/TestTfpSnapshotRestoreUpgradeStrategy$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/snapshot --junitfile results.xml -- -timeout=60m -v -run "TestTfpSnapshotRestoreUpgradeStrategyTestSuite/TestTfpSnapshotRestoreUpgradeStrategyDynamicInput$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.