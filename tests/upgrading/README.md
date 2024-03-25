# Upgrading

In the upgrading tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Upgrade the cluster's Kubernetes version to the desired version
4. Perform post-upgrading checks
7. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Upgrading Clusters](#Upgrading-Clusters)

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

## Upgrading Clusters
Similar to the `provisioning` tests, the node scaling tests have static test cases as well as dynamicInput tests you can specify. In order to run the dynamicInput tests, you will need to define the `terratest` block in your config file. See an example below:

```yaml
terratest:
  kubernetesVersion: ""
  upgradedKubernetesVersion: "" # If left blank or is omitted completely, the latest version in Rancher will be used. This is only for RKE1/RKE2/K3s. Hosted clusters MUST have this filled out.
  psact: "" # Optional, can be left out or can have values `rancher-privileged` or `rancher-restricted`
  ```

Additionally, you will need to ensure that the initial cluster version is NOT the latest version found in Rancher. This is because if you leave `upgradedKubernetesVersion` blank, then the test will automatically upgrade to the latest version found in Rancher.

See the below examples on how to run the tests:

### RKE1/RKE2/K3S

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/upgrading --junitfile results.xml -- -timeout=60m -v -run "TestTfpKubernetesUpgradeTestSuite/TestTfpKubernetesUpgrade$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/upgrading --junitfile results.xml -- -timeout=60m -v -run "TestTfpKubernetesUpgradeTestSuite/TestTfpKubernetesUpgradeDynamicInput$"`

### Hosted

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/upgrading --junitfile results.xml -- -timeout=60m -v -run "TestTfpKubernetesUpgradeHostedTestSuite/TestTfpKubernetesUpgradeHosted$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.