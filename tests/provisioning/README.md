# Provisioning

In the provisioning tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
7. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Provisioning Clusters](#Provisioning=Clusters)

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
  kubernetesVersion: "v1.26.10+k3s2"
  psact: "" # Optional, can be left out or can have values `rancher-privileged` or `rancher-restricted`
  ```

See the below examples on how to run the tests:

`go test -v -timeout 60m -run ^TestProvisionTestSuite/TestProvision$` \
`go test -v -timeout 60m -run ^TestProvisionTestSuite/TestProvisionDynamicInput$`