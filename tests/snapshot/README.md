# Snapshot

In the snapshot tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Check if the downstream cluster is RKE1, RKE2 or K3S
4. Perform etcd snapshot
5. Perform post etcd snapshot checks
6. Perform etcd restore
7. Perform post etcd restore checks
7. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

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
  kubernetesVersion: "v1.26.11+k3s2"
  upgradeKubernetesVersion: "" # If left blank, the default version in Rancher will be used.
  snapshotRestore: "all" # Options include none, kubernetesVersion, all. Option 'none' means that only the etcd will be restored.
  controlPlaneConcurrencyValue: "15%"
  workerConcurrencyValue: "20%"
  controlPlaneUnavailableValue: "1"
  workerUnavailableValue: "10%"
  recurringRestores: 1 # By default, this is set to 1 if this field is not included in the config.
  ```

See the below examples on how to run the tests:

### Snapshot restore
`go test -v -timeout 60m -run "TestSnapshotRestoreTestSuite/TestSnapshotRestore"` \
`go test -v -timeout 60m -run "TestSnapshotRestoreTestSuite/TestSnapshotRestoreDynamicInput"`

### Snapshot restore with K8s upgrade
`go test -v -timeout 60m -run "TestSnapshotRestoreK8sUpgradeTestSuite/TestSnapshotRestoreK8sUpgrade"` \
`go test -v -timeout 60m -run "TestSnapshotRestoreK8sUpgradeTestSuite/TestSnapshotRestoreK8sUpgradeDynamicInput"`

### Sanpshot restore with upgrade strategy
`go test -v -timeout 60m -run "TestSnapshotRestoreUpgradeStrategyTestSuite/TestSnapshotRestoreUpgradeStrategy"` \
`go test -v -timeout 60m -run "TestSnapshotRestoreUpgradeStrategyTestSuite/TestSnapshotRestoreUpgradeStrategyDynamicInput"`