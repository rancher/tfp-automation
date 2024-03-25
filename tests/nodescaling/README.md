# Node scaling

In the node scaling tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Scale the downstream cluster's nodes up to the desired amount
4. Perform post-scaling up checks
5. Scale the downstream cluster's nodes down to the desired amount
6. Perform post-scaling down checks
7. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Scaling Node Pools](#Scaling-Node-Pools)

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

## Scaling Node Pools
Similar to the `provisioning` tests, the node scaling tests have static test cases as well as dynamicInput tests you can specify. In order to run the dynamicInput tests, you will need to define the `terratest` block in your config file. See an example below:

### RKE1/RKE2/K3S

```yaml
terratest:
  kubernetesVersion: ""
  nodeCount: 3
  scaledUpNodeCount: 8
  scaledDownNodeCount: 6
  psact: "" # Optional, can be left out or can have values `rancher-privileged` or `rancher-restricted`
  nodepools:
    - etcd: true
      controlplane: false
      worker: false
      quantity: 1
    - etcd: false
      controlplane: true
      worker: false
      quantity: 1
    - etcd: false
      controlplane: false
      worker: true
      quantity: 1
  scalingInput:
    scaledUpNodepools:
      - etcd: true
        controlplane: false
        worker: false
        quantity: 3
      - etcd: false
        controlplane: true
        worker: false
        quantity: 2
      - etcd: false
        controlplane: false
        worker: true
        quantity: 3
    scaledDownNodepools:
      - etcd: true
        controlplane: false
        worker: false
        quantity: 3
      - etcd: false
        controlplane: true
        worker: false
        quantity: 2
      - etcd: false
        controlplane: false
        worker: true
        quantity: 1
  ```

### AKS

```yaml
terratest:
  kubernetesVersion: ""
  nodeCount: 3
  scaledUpNodeCount: 8
  scaledDownNodeCount: 6
  nodepools:
    - quantity: 3
  scalingInput:
    scaledUpNodepools:
      - quantity: 8
    scaledDownNodepools:
      - quantity: 6
```

### EKS

```yaml
terratest:
  kubernetesVersion: ""
  nodeCount: 3
  scaledUpNodeCount: 8
  scaledDownNodeCount: 6
  nodepools:
    - instanceType: "t3.medium"
      desiredSize: 3
      maxSize: 10
      minSize: 3
  scalingInput:
    scaledUpNodepools:
      - instanceType: "t3.medium"
        desiredSize: 8
        maxSize: 10
        minSize: 3
    scaledDownNodepools:
      - instanceType: "t3.medium"
        desiredSize: 6
        maxSize: 10
        minSize: 3
```

### GKE

```yaml
terratest:
  kubernetesVersion: ""
  nodeCount: 3
  scaledUpNodeCount: 8
  scaledDownNodeCount: 6
  nodepools:
    - quantity: 3
      maxPodsContraint: 110
  scalingInput:
    scaledUpNodepools:
      - quantity: 8
        maxPodsContraint: 110
    scaledDownNodepools:
      - quantity: 6
        maxPodsContraint: 110
```

See the below examples on how to run the tests:

### RKE1/RKE2/K3S

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/nodescaling --junitfile results.xml -- -timeout=60m -v -run "TestTfpScaleTestSuite/TestTfpScale$"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/nodescaling --junitfile results.xml -- -timeout=60m -v -run "TestTfpScaleTestSuite/TestTfpScaleDynamicInput$"`

### Hosted

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/nodescaling --junitfile results.xml -- -timeout=60m -v -run "TestTfpScaleHostedTestSuite/TestTfpScaleHosted$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.