# Airgap Recurring Tests

In the airgap recurring tests, there are various tests that are ran within an airgap environment. In general, here is the workflow followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Perform specific test functionality
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

As we are operating within an airgapped environment, you will not be able to connect in your browser without first connecting via a jump host. The easiest way to do this is with the following command: `ssh -i <PEM file> -f -N -L 8443:<Rancher FQDN:443 <username>@<Bastion public IP>`.

The above command will connect your client node to the same network that your airgapped environment, thus allowing you to access in your browser Additionally, in your client node's `/etc/hosts` file, temporarily update to have the following entry:

`127.0.0.1 <Rancher FQDN>`

Both the ssh command and the temporary `/etc/hosts` entry is needed in order to connect to the airgapped environment.

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

Reference the example config block below:

```yaml
terraform:
  airgapBastion: ""
  cni: ""
  enableNetworkPolicy: false
  defaultClusterRoleForProjectMembers: "user"
  downstreamClusterProvider: ""       # REQUIRED - can be aws, azure, linode, vsphere
  localAuthEndpoint: false      # OPTIONAL - false by default
  privateKeyPath: ""
  privateRegistries:                          # This is an optional block. You must already have a private registry stood up
    url: ""
    username: ""
    password: ""
    systemDefaultRegistry: ""
    insecure: true
    authConfigSecretName: ""
    mirrorHostname: ""
    mirrorEndpoint: ""
  
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
    - quantity: 1
      windows: true
```

See the below examples on how to run the tests:

### ACE
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/airgap --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestAirgapACETestSuite/TestTfpAirgapACE$"`

### API
`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/airgap --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestAirgapAPITestSuite/TestTfpAirgapAPI$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/provisioning --junitfile results.xml --jsonfile results.json -- -timeout=60m -tags=validation -v -run "TestAirgapACETestSuite/TestTfpAirgapACE$";/path/to/tfp-automation/reporter`