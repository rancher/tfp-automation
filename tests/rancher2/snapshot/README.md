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
2. [Local Qase Reporting](#Local-Qase-Reporting)

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
    s3:                               # Optional block, use if you want an S3 snapshot
      bucket: ""
      cloudCredentialName: ""
      endpoint: "s3.us-east-2.amazonaws.com"
      endpointCA: ""
      folder: ""
      region: "us-east-2"
      skipSSLVerify: true
terratest:
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
  snapshotInput: {}
```

To see what goes into the `terraform` block in addition to the `rancher`, please refer to the tfp-automation [README](../../README.md).

See the below examples on how to run the tests:

### Snapshot restore
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/snapshot --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpSnapshotRestoreTestSuite/TestTfpSnapshotRestore$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/snapshot --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=dynamic -v -run "TestTfpSnapshotRestoreTestSuite/TestTfpSnapshotRestoreDynamicInput$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/snapshot --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestTfpSnapshotRestoreTestSuite/TestTfpSnapshotRestore$";/path/to/tfp-automation/reporter`