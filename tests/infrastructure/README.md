# Setup Infrastructure

The `infrastructure` directory aims to be a hub to create various environments for testing needs. These include standalone clusters as well as various types of Rancher setups (i.e. airgap or proxy). These are written in the forms of tests, but they are strictly meant to get you going with an environment in a stable, reliable manner.

Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Setup Rancher](#Setup-Rancher)
2. [Setup Airgap Rancher](#Setup-Airgap-Rancher)
3. [Setup Proxy Rancher](#Setup-Proxy-Rancher)
4. [Setup RKE1 Cluster](#Setup-RKE1-Cluster)
5. [Setup RKE2 Cluster](#Setup-RKE2-Cluster)
6. [Setup Airgap RKE2 Cluster](#Setup-Airgap-RKE2-Cluster)
6. [Setup K3S Cluster](#Setup-K3S-Cluster)

## Setup Rancher

See below an example config on setting up a Rancher server powered by a RKE2 HA cluster:

```yaml
#######################
# RANCHER CONFIG
#######################
rancher:
  host: ""                                        # REQUIRED - fill out with the expected Rancher server URL
  insecure: true                                  # REQUIRED - leave this as true
#######################
# TERRAFORM CONFIG
#######################
terraform:
  cni: ""
  provider: ""                                # REQUIRED - supported values are aws | linode | harvester
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""
  ##########################################
  # STANDALONE CONFIG - INFRASTRUCTURE SETUP
  ##########################################
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
  linodeCredentials:
    linodeToken: ""  
  linodeConfig:
    clientConnThrottle: 20
    domain: ""
    linodeImage: ""
    linodeRootPass: ""
    privateIP: true
    region: ""
    soaEmail: ""
    swapSize: 256
    tags: [""]
    timeout: "5m"
    type: ""
  ###################################
  # STANDALONE CONFIG - RANCHER SETUP
  ###################################
  standalone:
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    rancherChartVersion: ""                       # REQUIRED - fill with desired value
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Group: ""                                 # REQUIRED - fill with group of the instance created
    rke2User: ""                                  # REQUIRED - fill with username of the instance created
    stagingRancherAgentImage: ""                  # OPTIONAL - fill out only if you are using staging registry
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value (i.e. v1.30.6+rke2r1)
```

Note: Depending on what `provider` is set to, only fill out the appropriate section. Before running locally, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export AWS_PROVIDER_VERSION=""
export LINODE_PROVIDER_VERSION=""
```

See the below examples on how to run the test:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=2h -v -run "TestRancherTestSuite$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Setup Airgap Rancher

See below an example config on setting up an air-gapped Rancher server powered by an air-gapped RKE2 HA cluster:

```yaml
terraform:
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsSecurityGroups: [""]
    awsSecurityGroupNames: [""]
    awsUser: ""
    region: ""
    registryRootSize: 500
    sshConnectionType: "ssh"
    timeout: "5m"
  standalone:
    airgapInternalFQDN: ""                        # REQUIRED - Have the same name as the rancherHostname but it must end with `-internal`
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    osGroup: ""                                   # REQUIRED - fill with group of the instance created
    osUser: ""                                    # REQUIRED - fill with username of the instance created
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using a custom registry
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Version: ""                               # REQUIRED - the format MUST be in `v1.xx.x` (i.e. v1.31.3)
  standaloneRegistry:
    assetsPath: ""                                # REQUIRED - ensure that you end with a trailing `/`
    authenticated: true                           # REQUIRED - true if you want an authenticated registry, false for a non-authenticated registry
    registryName: ""                              # REQUIRED (authenticated registry only)
    registryPassword: ""                          # REQUIRED (authenticated registry only)
    registryUsername: ""                          # REQUIRED (authenticated registry only)
```

Before running, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export AWS_PROVIDER_VERSION=""
```

See the below examples on how to run the test:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=7h -v -run "TestAirgapRancherTestSuite$"`

## Setup Proxy Rancher

See below an example config on setting up a Rancher server behind a proxy, powered by an RKE2 HA cluster:

```yaml
terraform:
  privateKeyPath: ""                              # REQUIRED - specify private key that will be used to access created instances
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsSecurityGroups: [""]
    awsRootSize: 100
    awsRoute53Zone: ""
    region: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
  standalone:
    bootstrapPassword: ""                         # REQUIRED - this is the same as the adminPassword above, make sure they match
    certManagerVersion: ""                        # REQUIRED - (e.g. v1.15.3)
    osGroup: ""                                   # REQUIRED - fill with desired value
    osUser: ""                                    # REQUIRED - fill with desired value
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using a custom registry
    rancherChartVersion: ""                       # REQUIRED - fill with desired value
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Group: ""                                 # REQUIRED - fill with group of the instance created
    rke2User: ""                                  # REQUIRED - fill with username of the instance created
    rke2Version: ""                               # REQUIRED - fill with desired RKE2 k8s value (i.e. v1.30.6+rke2r1)
  standaloneRegistry:
    registryName: ""                              # REQUIRED - fill with desired value
    registryPassword: ""                          # REQUIRED - fill with desired value
    registryUsername: ""                          # REQUIRED - fill with desired value
```

Before running, be sure to run the following commands:

```yaml
export RANCHER2_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export LOCALS_PROVIDER_VERSION=""
export AWS_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=3h -v -run "TestProxyRancherTestSuite$"`

## Setup RKE1 Cluster

See below an example config on setting up a standalone RKE1 cluster:

```yaml
terraform:
  privateKeyPath: ""
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    region: ""
    awsUser: ""
    sshConnectionType: ""
    standaloneSecurityGroupNames: [""]
    timeout: ""
  standalone:
    osUser: ""
```

Before running, be sure to run the following commands:

```yaml
export RKE_PROVIDER_VERSION=""
export CATTLE_TEST_CONFIG=<path/to/yaml>
export AWS_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestRKEProviderTestSuite$"`

## Setup RKE2 Cluster

See below an example config on setting up a standalone RKE2 cluster:

```yaml
terraform:
  cni: ""
  provider: ""                                # REQUIRED - supported values are aws | linode | harvester
  privateKeyPath: ""
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroups: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: ""
  linodeCredentials:
    linodeToken: ""  
  linodeConfig:
    clientConnThrottle: 20
    domain: ""
    linodeImage: ""
    linodeRootPass: ""
    privateIP: true
    region: ""
    soaEmail: ""
    swapSize: 256
    tags: [""]
    timeout: "5m"
    type: ""
  standalone:
    osGroup: ""                                   # REQUIRED - fill with group of the instance created
    osUser: ""                                    # REQUIRED - fill with username of the instance created
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using a custom registry
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Version: ""                               # REQUIRED - the format MUST be in `v1.xx.x` (i.e. v1.31.3)
  standaloneRegistry:
    assetsPath: ""                                # REQUIRED - ensure that you end with a trailing `/`
    registryName: ""                              # REQUIRED - fill with desired value
```

Before running, be sure to run the following commands:

```yaml
export CATTLE_TEST_CONFIG=<path/to/yaml>
export AWS_PROVIDER_VERSION=""
export LINODE_PROVIDER_VERSION=""
export LOCALS_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestCreateRKE2ClusterTestSuite$"`

## Setup Airgap RKE2 Cluster

See below an example config on setting up a standalone airgapped RKE2 cluster:

```yaml
terraform:
  privateKeyPath: ""
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroups: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsUser: ""
    registryRootSize: 500
    sshConnectionType: "ssh"
    timeout: ""
  standalone:
    airgapInternalFQDN: ""                        # REQUIRED - Have the same name as the rancherHostname but it must end with `-internal`
    osGroup: ""                                   # REQUIRED - fill with group of the instance created
    osUser: ""                                    # REQUIRED - fill with username of the instance created
    rancherAgentImage: ""                         # OPTIONAL - fill out only if you are using a custom registry
    rancherChartRepository: ""                    # REQUIRED - fill with desired value. Must end with a trailing /
    rancherHostname: ""                           # REQUIRED - fill with desired value
    rancherImage: ""                              # REQUIRED - fill with desired value
    rancherTagVersion: ""                         # REQUIRED - fill with desired value
    repo: ""                                      # REQUIRED - fill with desired value
    rke2Version: ""                               # REQUIRED - the format MUST be in `v1.xx.x` (i.e. v1.31.3)
  standaloneRegistry:
    assetsPath: ""                                # REQUIRED - ensure that you end with a trailing `/`
    registryName: ""                              # REQUIRED - fill with desired value
```

Before running, be sure to run the following commands:

```yaml
export CATTLE_TEST_CONFIG=<path/to/yaml>
export AWS_PROVIDER_VERSION=""
export LOCALS_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=5h -v -run "TestCreateAirgappedRKE2ClusterTestSuite$"`

## Setup K3S Cluster

See below an example config on setting up a standalone K3S cluster:

```yaml
terraform:
  cni: ""
  provider: ""                                # REQUIRED - supported values are aws | linode | harvester
  privateKeyPath: ""
  resourcePrefix: ""
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroups: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: ""
    awsRootSize: 100
    awsRoute53Zone: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: ""
  linodeCredentials:
    linodeToken: ""  
  linodeConfig:
    clientConnThrottle: 20
    domain: ""
    linodeImage: ""
    linodeRootPass: ""
    privateIP: true
    region: ""
    soaEmail: ""
    swapSize: 256
    tags: [""]
    timeout: "5m"
    type: ""
  standalone:
    k3sVersion: ""                                # REQUIRED - the format MUST be in `v1.xx.x+k3s1` (i.e. v1.31.4+k3s1)
    osGroup: ""                                   # REQUIRED - fill with group of the instance created
    osUser: ""                                    # REQUIRED - fill with username of the instance created
    rancherHostname: ""                           # REQUIRED - fill with desired value
```

Before running, be sure to run the following commands:

```yaml
export CATTLE_TEST_CONFIG=<path/to/yaml>
export AWS_PROVIDER_VERSION=""
export LINODE_PROVIDER_VERSION=""
export LOCALS_PROVIDER_VERSION=""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/infrastructure --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestCreateK3SClusterTestSuite$"`