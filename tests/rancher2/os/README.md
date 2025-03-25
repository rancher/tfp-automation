# OS Checks

In the OS check test, the following workflow is followed:

1. Provision all permutations for a given AMI in parallel
2. Run all the workload tests
3. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)
4. Repeat steps 1-3 for all AMIs provided

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).


```yaml
rancher:
  host: ""
  adminToken: ""
  insecure: true
  cleanup: true

terraform:
  module: [ec2_rke1, ec2_rke2, ec2_k3s, ec2_rke1_custom, ec2_rke2_custom, ec2_k3s_custom, ec2_rke1_import, ec2_rke2_import, ec2_k3s_import]
  cni: [calico]
  resourcePrefix: "oscheck"
  privateKeyPath: ""

  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  
  awsConfig:
    awsUser: "ec2-user"
    ami: [""]
    awsInstanceType: t3a.medium
    region: us-east-2
    awsSecurityGroupNames: [""]
    awsSecurityGroups: [""]
    awsVpcID: 
    awsZoneLetter: a
    awsRootSize: 100
    awsKeyName: ""

  #Standalone is for import Clusters
    standaloneSecurityGroupNames: [""]
    sshConnectionType: "ssh"
    timeout: "5m"
  standalone:
    rke2Version: "v1.32.2+rke2r1"
    k3sVersion: "v1.32.2+k3s1"
    osGroup: "docker"
    osUser: "ec2-user"

```yaml
terratest:
  nodepools:
    - quantity: 1
      etcd: true
      controlplane: false
      worker: false
    - quantity: 1
      etcd: false
      controlplane: true
      worker: false
    - quantity: 1
      etcd: false
      controlplane: false
      worker: true
  kubernetesVersion: [v1.32.2-rancher1-1, v1.32.2+rke2r1, v1.32.2+k3s1]
  nodeCount: 3
```

### Run Command:
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/os --junitfile results.xml -- -tags=validation -run "TestOSValidationTestSuite/TestDynamicOSValidation" -timeout=4h -v`

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/os --junitfile results.xml -- -tags=validation -run "TestOSValidationTestSuite/TestDynamicOSValidation" -timeout=4h -v;/path/to/tfp-automation/reporter`