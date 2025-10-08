# Recurring Runs

The `recurring` directory will spin up a Rancher environment, with the desired node provider, and will run various Rancher2 tests that are used for release testing. This is ran on a weekly basis alongside various infrastrucre-based testing via Github Actions.

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Getting Started](#Getting-Started)
2. [Local Qase Reporting](#Local-Qase-Reporting)

## Getting Started

See below an example config on setting up a Rancher server powered by a RKE2 HA cluster using AWS as the node provider:

```yaml
rancher:
  host: ""                                        # REQUIRED - fill out with the expected Rancher server URL
  insecure: true                                  # REQUIRED - leave this as true

upgradeInput:
  clusters:
    -  versionToUpgrade: ""                       # Leave off the suffix; the test will add it for K3s and RKE2

terraform:
  cni: ""
  provider: ""                                    # REQUIRED - supported values are aws | linode | harvester | vsphere
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  windowsPrivateKeyPath: ""
  resourcePrefix: ""
  privateRegistries:                              # This is an optional block. You must already have a private registry stood up
    engineInsecureRegistry: ""                    # RKE1 specific
    url: ""
    systemDefaultRegistry: ""                     # RKE2/K3S specific, can be left blank
    username: ""                                  # RKE1 specific
    password: ""                                  # RKE1 specific
    insecure: true
    authConfigSecretName: ""                      # RKE2/K3S specific
    mirrorHostname: ""
    mirrorEndpoint: ""
    mirrorRewrite: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    awsSecurityGroups: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    region: ""
    prefix: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    windowsAMI2019: ""
    windowsAMI2022: ""
    windowsAWSUser: ""
    windows2019Password: ""
    windows2022Password: ""
    windowsInstanceType: ""
    windowsKeyName: ""
    ipAddressType: "ipv4"
    loadBalancerType: "ipv4"
    targetType: "instance"

  standalone:
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    k3sVersion: ""                                # REQUIRED - fill with desired K3s k8s value (make sure it's not the highest version)
    rancherChartVersion: ""                       # REQUIRED - fill with desired value
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    registryPassword: ""                          # REQUIRED
    registryUsername: ""                          # REQUIRED
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Group: ""                                 # REQUIRED - fill with group of the instance created
    rke2User: ""                                  # REQUIRED - fill with username of the instance created
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using staging registry
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value (make sure it's not the highest version)

terratest:
  etcdCount: 3
  controlPlaneCount: 2
  workerCount: 3
  windowsNodeCount: 1
  pathToRepo: "go/src/github.com/rancher/tfp-automation"
  snapshotInput: {}
```

Note: Depending on what `provider` is set to, only fill out the appropriate section. Before running locally, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export CLOUD_PROVIDER_VERSION=""
```

See the below examples on how to run the test:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/recurring --junitfile results.xml --jsonfile results.json -- -timeout=3h -v -run "TfpRancher2RecurringRunsTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/recurring --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TfpRancher2RecurringRunsTestSuite$";/path/to/tfp-automation/reporter`