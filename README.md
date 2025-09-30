# <p align="center">RANCHER :small_blue_diamond: TFP-AUTOMATION</p>

[![Airgap](https://github.com/rancher/tfp-automation/actions/workflows/airgap-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/airgap-test.yaml)
[![Airgap Upgrade](https://github.com/rancher/tfp-automation/actions/workflows/airgap-upgrade-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/airgap-upgrade-test.yaml)

[![Post Release Sanity](https://github.com/rancher/tfp-automation/actions/workflows/post-release-sanity-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/post-release-sanity.yaml)

[![Post Release Upgrade Sanity](https://github.com/rancher/tfp-automation/actions/workflows/post-release-sanity-upgrade-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/post-release-sanity-upgrade.yaml)

[![Proxy](https://github.com/rancher/tfp-automation/actions/workflows/proxy-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/proxy-test.yaml)
[![Proxy ARM64](https://github.com/rancher/tfp-automation/actions/workflows/proxy-arm64-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/proxy-arm64-test.yaml)
[![Proxy Upgrade](https://github.com/rancher/tfp-automation/actions/workflows/proxy-upgrade-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/proxy-upgrade-test.yaml)
[![Proxy Upgrade ARM64](https://github.com/rancher/tfp-automation/actions/workflows/proxy-upgrade-arm64-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/proxy-upgrade-arm64-test.yaml)

[![Rancher2 Recurring](https://github.com/rancher/tfp-automation/actions/workflows/rancher2-recurring-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/rancher2-recurring-test.yaml)

[![Registry](https://github.com/rancher/tfp-automation/actions/workflows/registry-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/registry-test.yaml)

[![Sanity](https://github.com/rancher/tfp-automation/actions/workflows/sanity-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/sanity-test.yaml)
[![Sanity ARM64](https://github.com/rancher/tfp-automation/actions/workflows/sanity-arm64-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/sanity-arm64-test.yaml)
[![Sanity Upgrade](https://github.com/rancher/tfp-automation/actions/workflows/sanity-upgrade-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/sanity-upgrade-test.yaml)
[![Sanity Upgrade ARM64](https://github.com/rancher/tfp-automation/actions/workflows/sanity-upgrade-arm64-test.yaml/badge.svg?branch=main)](https://github.com/rancher/tfp-automation/actions/workflows/sanity-upgrade-arm64-test.yaml)


`tfp-automation` is a Github Actions based testing framework designed to handle the following tasks:
- Conduct daily regression testing amongst supported Rancher release lines
- Automate release testing across different permutations of a Rancher HA environment (e.g. normal, airgap, proxy)
- Support infrastructure creation for various node providers

The above points are done with an emphasis on testing the Rancher2 Terraform provider. This framework utilizes Terratest alongside Go to accomplish these goals.

---

<a name="top"></a>

# <p align="center"> :scroll: Table of contents </p>

-   [Configurations](#configurations)
    -   [Infrastructure](#configurations-infrastructure)
    -   [Rancher](#configurations-rancher)
    -   [Terraform](#configurations-terraform)
        -   [AKS](#configurations-terraform-aks)
        -   [EKS](#configurations-terraform-eks)
        -   [GKE](#configurations-terraform-gke)
        -   [AZURE_RKE1](#configurations-terraform-azure_rke1)
        -   [EC2_RKE1](#configurations-terraform-ec2_rke1)
        -   [HARVESTER_RKE1](#configurations-terraform-harvester_rke1)
        -   [LINODE_RKE1](#configurations-terraform-linode_rke1)
        -   [VSPHERE_RKE1](#configurations-terraform-vsphere_rke1)
        -   [AZURE_RKE2 + AZURE_K3S](#configurations-terraform-rke2_k3s_azure)
        -   [EC2_RKE2 + EC2_K3S](#configurations-terraform-rke2_k3s_ec2)
        -   [HARVESTER_RKE2 + HARVESTER_K3S](#configurations-terraform-rke2_k3s_harvester)
        -   [LINODE_RKE2 + LINODE_K3S](#configurations-terraform-rke2_k3s_linode)
        -   [VSPHERE_RKE2 + VSPHERE_K3S](#configurations-terraform-rke2_k3s_vsphere)
    -   [Terratest](#configurations-terratest)
        -   [Nodepools](#configurations-terratest-nodepools)
            -   [AKS Nodepools](#configurations-terratest-nodepools-aks)
            -   [EKS Nodepools](#configurations-terratest-nodepools-eks)
            -   [GKE Nodepools](#configurations-terratest-nodepools-gke)
            -   [RKE1, RKE2, K3S Nodepools](#configurations-terratest-nodepools-rke1_rke2_k3s)
        -  [Provision](#configurations-terratest-provision)
        -  [Scale](#configurations-terratest-scale)
        -  [Kubernetes Upgrade](#configurations-terratest-kubernetes_upgrade)
        -  [Snapshots](#configurations-terratest-snapshots)
        -  [Build Module](#configurations-terratest-build_module)
        -  [Cleanup](#configurations-terratest-cleanup)

---

<a name="configurations"></a>

### <p align="center"> Configurations </p>

##### When testing locally, the following environment variables should be exported:
```yaml
export RANCHER2_PROVIDER_VERSION=""                                     # Required
export CLOUD_PROVIDER_VERSION=""                                        # Required for custom cluster / infrastructure building
export KUBERNETES_PROVIDER_VERSION=""                                   # Required for infrastructure building using Harvester
export LOCALS_PROVIDER_VERSION=""                                       # Required for custom cluster / infrastructure building
export QASE_AUTOMATION_TOKEN=""                                         # Required for local Qase reporting
export QASE_TEST_RUN_ID=""                                              # Required for local Qase reporting
```
##### These tests require an accurately configured `cattle-config.yaml` to successfully run.

##### Each `cattle-config.yaml` must include the following configurations:

```yaml
rancher:
  # define rancher specific configs here

terraform:
  # define module specific configs here
  
terratest:
  # define test specific configs here
```

---

<a name="configurations-infrastructure"></a>
#### :small_red_triangle: [Back to top](#top)

As part of this framework, you have the ability to spin up a Rancher HA environment using various node providers. The `infrasturcture` folder hosts a series of different test files that explain further. As far as different configurations needed to make this happen, see below:

##### Infrastructure

```yaml
terraform:
  cni: "calico"
  defaultClusterRoleForProjectMembers: "true"
  enableNetworkPolicy: false
  provider: ""                              # The following providers are supported: aws | linode | harvester
  privateKeyPath: ""
  resourcePrefix: ""
  windowsPrivateKeyPath: ""

  # Fill out the AWS section if provider is set to aws.
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami: ""
    awsKeyName: ""
    awsInstanceType: ""
    region: ""
    awsSecurityGroups: [""]
    awsSecurityGroupNames: [""]
    awsSubnetID: ""
    awsVpcID: ""
    awsZoneLetter: "a"
    awsRootSize: 100
    awsRoute53Zone: ""
    awsUser: ""
    sshConnectionType: "ssh"
    timeout: "5m"
    windowsAMI: ""
    windowsAwsUser: ""
    windowsInstanceType: ""
    windowsKeyName: ""

  # Fill out the Linode section if provider is set to linode.
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
  
  # Fill out this Harvester section if provider is set to harvester.
  harvesterCredentials:
    clusterId: ""
    clusterType: "imported"
    kubeconfigContent: ""
  harvesterConfig:
    diskSize: "30"
    cpuCount: "4"
    memorySize: "8"
    networkNames: [""]
    imageName: ""
    vmNamespace: "default"
    sshUser: ""

  # Fill out this vSphere section if provider is set to vsphere.
  vsphereCredentials:
    password: ""
    username: ""
    vcenter: ""
  vsphereConfig:  
    cloneFrom: ""
    cpuCount: ""
    datacenter: ""
    datastore: ""
    datastoreCluster: ""
    diskSize: ""
    guestID: "ubuntu64Guest"      # This will change depending on the OS you're using
    folder: ""
    hostSystem: ""
    memorySize: ""
    standaloneNetwork: ""
    vsphereUser: ""

  standalone:
    bootstrapPassword: ""
    certManagerVersion: "v1.15.3"
    osUser: ""
    osGroup: ""
    rancherChartRepository: "https://releases.rancher.com/server-charts/"
    rancherHostname: ""
    rancherImage: "rancher/rancher"
    rancherTagVersion: "v2.11.0"
    repo: "latest"
    rke2Version: "v1.30.9+rke2r1"
```

---

<a name="configurations-rancher"></a>
#### :small_red_triangle: [Back to top](#top)

The `rancher` configurations in the `cattle-config.yaml` will remain consistent across all modules and tests.

##### Rancher

```yaml
rancher:
  host: url-to-rancher-server.com
  adminToken: token-XXXXX:XXXXXXXXXXXXXXX
  insecure: true
  cleanup: true
```

---

<a name="configurations-terraform"></a>
#### :small_red_triangle: [Back to top](#top)

The `terraform` configurations in the `cattle-config.yaml` are module specific.  Fields to configure vary per module. Below are generic fields that are applicable regardless of module. See them below:

##### Terraform

```yaml
terraform:
  etcd:                                       # This is an optional block.
    disableSnapshot: false
    snapshotCron: "0 */5 * * *"
    snapshotRetention: 6
    s3:
      bucket: ""
      cloudCredentialName: ""
      endpoint: ""
      endpointCA: ""
      folder: ""
      region: ""
      skipSSLVerify: true
  etcdRKE1:                                   # This is an optional block
    backupConfig:
      enabled: true
      intervalHours: 12
      safeTimestamp: true
      timeout: 120
      s3BackupConfig:
        accessKey: ""
        bucketName: ""
        endpoint: ""
        folder: ""
        region: ""
        secretKey: ""
    retention: "72h"
    snapshot: false
  cloudCredentialName: ""
  defaultClusterRoleForProjectMembers: "true" # Can be "true" or "false"
  enableNetworkPolicy: false                  # Can be true or false
  hostnamePrefix: ""   
  machineConfigName: ""                       # RKE2/K3S specific
  networkPlugin: ""                           # RKE1 specific
  nodeTemplateName: ""                        # RKE1 specific
  privateRegistries:                          # This is an optional block. You must already have a private registry stood up
    engineInsecureRegistry: ""                # RKE1 specific
    url: ""
    systemDefaultRegistry: ""                 # RKE2/K3S specific, can be left blank
    username: ""                              # RKE1 specific
    password: ""                              # RKE1 specific
    insecure: true
    authConfigSecretName: ""                  # RKE2/K3S specific
    mirrorHostname: ""
    mirrorEndpoint: ""
    mirrorRewrite: ""
  chartValues: |-			      # Provided as a multiline string
    rke2-cilium:                              # RKE2/Cilium specific example of how to do a Kube-proxy Replacement deployment
      k8sServiceHost: 127.0.0.1               
      k8sServicePort: 6443
      kubeProxyReplacement: true
  cni: cilium				      # RKE2 specific
  disable-kube-proxy: true		      # Can be "true" or "false"
```

Note: At this time, private registries for RKE2/K3s MUST be used with provider version 3.1.1. This is due to issue https://github.com/rancher/terraform-provider-rancher2/issues/1305.

<a name="configurations-terraform-aks"></a>
#### :small_red_triangle: [Back to top](#top)

###### AKS

```yaml
terraform:
  module: aks
  cloudCredentialName: tf-aks
  azureCredentials:
    clientId: ""
    clientSecret: ""
    environment: "AzurePublicCloud"
    subscriptionId: ""
    tenantId: ""
  azureConfig:
    availabilityZones:
      - '1'
      - '2'
      - '3'
    image: ""
    location: ""
    managedDisks: false
    mode: "System"
    name: "agentpool"
    networkDNSServiceIP: ""
    networkDockerBridgeCIDR: ""
    networkServiceCIDR: ""
    noPublicIp: false
    osDiskSizeGB: 128
    outboundType: "loadBalancer"
    resourceGroup: ""
    resourceLocation: ""
    subnet: ""
    taints: ["none:PreferNoSchedule"]
    vmSize: Standard_DS2_v2
    vnet: ""
    tfLogging: false
```

---

<a name="configurations-terraform-eks"></a>
#### :small_red_triangle: [Back to top](#top)

###### EKS

```yaml
terraform:
  module: eks
  cloudCredentialName: tf-eks
  hostnamePrefix: tfp
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    awsInstanceType: t3.medium
    region: us-east-2
    awsSubnets:
      - ""
      - ""
    awsSecurityGroups:
      - ""
    publicAccess: true
    privateAccess: true
```

---


<a name="configurations-terraform-gke"></a>
#### :small_red_triangle: [Back to top](#top)

###### GKE

```yaml
terraform:
  module: gke
  cloudCredentialName: tf-creds-gke
  hostnamePrefix: tfp
  googleCredentials:
    authEncodedJson: |-
      {
        "type": "service_account",
        "project_id": "",
        "private_key_id": "",
        "private_key": "",
        "client_email": "",
        "client_id": "",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_x509_cert_url": ""
      }
  googleConfig:
    region: us-central1-c
    projectID: ""
    network: default
    subnetwork: default
```

---

<a name="configurations-terraform-azure_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

###### AZURE_RKE1

```yaml
terraform:
  module: azure_rke1
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  azureCredentials:
    clientId: ""
    clientSecret: ""
    environment: "AzurePublicCloud"
    subscriptionId: ""
    tenantId: ""
  azureConfig:
    availabilitySet: "docker-machine"
    subscriptionId: ""
    customData: ""
    diskSize: "100"
    dockerPort: "2376"
    faultDomainCount: "3"
    image: "Canonical:0001-com-ubuntu-server-jammy:22_04-lts:latest"
    location: "westus2"
    managedDisks: false
    noPublicIp: false
    openPort: ["6443/tcp","2379/tcp","2380/tcp","8472/udp","4789/udp","9796/tcp","10256/tcp","10250/tcp","10251/tcp","10252/tcp"]
    privateIpAddress: ""
    resourceGroup: ""
    size: "Standard_D2_v2"
    sshUser: "azureuser"
    staticPublicIp: false
    storageType: "Standard_LRS"
    updateDomainCount: "5"
```
---

<a name="configurations-terraform-ec2_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

###### EC2_RKE1

```yaml
terraform:
  module: ec2_rke1
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami:
    awsInstanceType: t3.medium
    region: us-east-2
    awsSecurityGroupNames:
      - security-group-name
    awsSubnetID: subnet-xxxxxxxx
    awsVpcID: vpc-xxxxxxxx
    awsZoneLetter: a
    awsRootSize: 80
```
---

<a name="configurations-terraform-harvester_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

###### HARVESTER_RKE1

```yaml
terraform:
  module: harvester_rke1
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  harvesterCredentials:
    clusterId: "c-m-clusterID"
    clusterType: "imported"
    kubeconfigContent: |
      kubeconfig-content
  harvesterConfig:
    diskSize: "30"
    cpuCount: "4"
    memorySize: "8"
    networkNames: ["default/net-name"]
    imageName: "default/image-name"
    vmNamespace: "default"
    sshUser: "ubuntu"
```
---

<a name="configurations-terraform-linode_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

###### LINODE_RKE1

```yaml
terraform:
  module: linode_rke1
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  linodeCredentials:
    linodeToken: ""
  linodeConfig:
    region: us-east
    linodeRootPass: ""
```
---

<a name="configurations-terraform-vsphere_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

###### VSPHERE_RKE1

```yaml
terraform:
  module: vsphere_rke1
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  vsphereCredentials:
    password: ""
    username: ""
    vcenter: ""
    vcenterPort: "443"
  vsphereConfig:  
    cfgparam: ["disk.enableUUID=TRUE"]
    cloneFrom: ""
    cloudConfig: ""
    contentLibrary: ""
    cpuCount: "4"
    creationType: "template"
    datacenter: ""
    datastore: ""
    datastoreCluster: ""
    diskSize: "40000"
    folder: ""
    hostsystem: ""
    memorySize: "8192"
    network: [""]
    pool: ""
    sshPassword: "tcuser"
    sshPort: "22"
    sshUser: "docker"
    sshUserGroup: "staff"
```

---

<a name="configurations-terraform-rke2_k3s_azure"></a>
#### :small_red_triangle: [Back to top](#top)

###### AZURE_RKE2 + AZURE_K3S

```yaml
terraform:
  module: azure_k3s
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  azureCredentials:
    clientId: ""
    clientSecret: ""
    environment: "AzurePublicCloud"
    subscriptionId: ""
    tenantId: ""
  azureConfig:
    availabilitySet: "docker-machine"
    customData: ""
    diskSize: "100"
    dockerPort: "2376"
    faultDomainCount: "3"
    image: "Canonical:0001-com-ubuntu-server-jammy:22_04-lts:latest"
    location: "westus2"
    managedDisks: false
    noPublicIp: false
    openPort: ["6443/tcp","2379/tcp","2380/tcp","8472/udp","4789/udp","9796/tcp","10256/tcp","10250/tcp","10251/tcp","10252/tcp"]
    privateIpAddress: ""
    resourceGroup: ""
    size: "Standard_D2_v2"
    sshUser: ""
    staticPublicIp: false
    storageType: "Standard_LRS"
    updateDomainCount: "5"
```

---

<a name="configurations-terraform-rke2_k3s_ec2"></a>
#### :small_red_triangle: [Back to top](#top)

###### EC2_RKE2 + EC2_K3S

```yaml
terraform:
  module: ec2_rke2
  cloudCredentialName: tf-creds-rke2
  machineConfigName: tf-rke2
  enableNetworkPolicy: false
  defaultClusterRoleForProjectMembers: user
  awsCredentials:
    awsAccessKey: ""
    awsSecretKey: ""
  awsConfig:
    ami:
    region: us-east-2
    awsSecurityGroupNames:
      - my-security-group
    awsSubnetID: subnet-xxxxxxxx
    awsVpcID: vpc-xxxxxxxx
    awsZoneLetter: a
```
---

<a name="configurations-terraform-rke2_k3s_harvester"></a>
#### :small_red_triangle: [Back to top](#top)

###### HARVESTER_RKE2 + HARVESTER_K3S

```yaml
terraform:
  module: harvester_rke2
  hostnamePrefix: tfp
  machineConfigName: tf-hvst
  harvesterCredentials:
    clusterId: "c-m-clusterID"
    clusterType: "imported"
    kubeconfigContent: |
      kubeconfig-content
  harvesterConfig:
    diskSize: "30"
    cpuCount: "4"
    memorySize: "8"
    networkNames: ["default/net-name"]
    imageName: "default/image-name"
    vmNamespace: "default"
    sshUser: "ubuntu"
```
---

<a name="configurations-terraform-rke2_k3s_linode"></a>
#### :small_red_triangle: [Back to top](#top)

###### LINODE_RKE2 + LINODE_K3S

```yaml
terraform:
  module: linode_k3s
  cloudCredentialName: tf-linode-creds
  machineConfigName: tf-k3s
  enableNetworkPolicy: false
  defaultClusterRoleForProjectMembers: user
  linodeCredentials:
    linodeToken: ""
  linodeConfig:
    linodeImage: linode/ubuntu20.04
    region: us-east
    linodeRootPass: xxxxxxxxxxxx
```
---

<a name="configurations-terraform-rke2_k3s_vsphere"></a>
#### :small_red_triangle: [Back to top](#top)

###### VSPHERE_RKE2 + VSPHERE_K3S

```yaml
terraform:
  module: vsphere_k3s
  networkPlugin: canal
  nodeTemplateName: tf-rke1-template
  hostnamePrefix: tfp
  vsphereCredentials:
    password: ""
    username: ""
    vcenter: ""
    vcenterPort: ""
  vsphereConfig:  
    cfgparam: ["disk.enableUUID=TRUE"]
    cloneFrom: ""
    cloudConfig: ""
    contentLibrary: ""
    cpuCount: "4"
    creationType: "template"
    datacenter: ""
    datastore: ""
    datastoreCluster: ""
    diskSize: "40000"
    folder: ""
    hostsystem: ""
    memorySize: "8192"
    network: [""]
    pool: ""
    sshPassword: "tcuser"
    sshPort: "22"
    sshUser: "docker"
    sshUserGroup: "staff"
```

---


<a name="configurations-terratest"></a>
#### :small_red_triangle: [Back to top](#top)

The `terratest` configurations in the `cattle-config.yaml` are test specific. Fields to configure vary per test. The `nodepools` field in the below configurations will vary depending on the module.  I will outline what each module expects first, then proceed to show the whole test specific configurations. 


<a name="configurations-terratest-nodepools"></a>
#### :small_red_triangle: [Back to top](#top)

###### Nodepools 
type: []Nodepool

<a name="configurations-terratest-nodepools-aks"></a>
#### :small_red_triangle: [Back to top](#top)

###### AKS Nodepool

AKS nodepools only need the `quantity` of nodes per pool to be provided, of type `int64`.  The below example will create a cluster with three node pools, each with a single node.

###### Example:
```yaml
nodepools:
  - quantity: 1
  - quantity: 1
  - quantity: 1
```

<a name="configurations-terratest-nodepools-eks"></a>
#### :small_red_triangle: [Back to top](#top)

###### EKS Nodepool

EKS nodepools require the `instanceType`, as type `string`, the `desiredSize` of the nodepool, as type `int64`, the `maxSize` of the node pool, as type `int64`, and the `minSize` of the node pool, as type `int64`. The minimum requirement for an EKS nodepool's `desiredSize` is `2`.  This must be respected or the cluster will fail to provision.

###### Example:
```yaml
nodepools:
  - instanceType: t3.medium
    desiredSize: 3
    maxSize: 3
    minSize: 0
```

<a name="configurations-terratest-nodepools-gke"></a>
#### :small_red_triangle: [Back to top](#top)

###### GKE Nodepool

GKE nodepools require the `quantity` of the node pool, as type `int64`, and the `maxPodsContraint`, as type `int64`.

###### Example:
```yaml
nodepools:
  - quantity: 2
    maxPodsContraint: 110
```

<a name="configurations-terratest-nodepools-rke1_rke2_k3s"></a>
#### :small_red_triangle: [Back to top](#top)

###### RKE1, RKE2, and K3S - all share the same nodepool configurations

For these modules, the required nodepool fields are the `quantity`, as type `int64`, as well as the roles to be assigned, each to be set or toggled via boolean - [`etcd`, `controlplane`, `worker`]. The following example will create three node pools, each with individual roles, and one node per pool.

###### Example:
```yaml
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
```

That wraps up the sub-section on nodepools, circling back to the test specific configs now...

Test specific fields to configure in this section are as follows:

<a name="configurations-terratest-provision"></a>
#### :small_red_triangle: [Back to top](#top)

##### Provision

```yaml
terratest:
  pathToRepo: # REQUIRED - path to repo from user's go directory i.e. ../go/<path/to/repo/tfp-automation>
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
  kubernetesVersion: ""
  nodeCount: 3

  # Below are the expected formats for all module kubernetes versions
  
  # AKS without leading v
  # e.g. '1.28.5'
  
  # EKS without leading v or any tail ending
  # e.g. '1.28'
  
  # GKE without leading v but with tail ending included
  # e.g. 1.28.7-gke.1026000
  
  # RKE1 with leading v and -rancher1-1 tail
  # e.g. v1.28.7-rancher1-1

  # RKE2 with leading v and +rke2r# tail
  # e.g. v1.28.7+rke2r1

  # K3S with leading v and +k3s# tail
  # e.g. v1.28.7+k3s1
```

Note: In this test suite, Terraform explicitly cleans up resources after each test case is performed. This is because Terraform will experience caching issues, causing tests to fail.

---

<a name="configurations-terratest-scale"></a>
#### :small_red_triangle: [Back to top](#top)

##### Scale

```yaml
terratest:
  pathToRepo: # REQUIRED - path to repo from user's go directory i.e. ../go/<path/to/repo/tfp-automation>
  kubernetesVersion: ""
  nodeCount: 3
  scaledUpNodeCount: 8
  scaledDownNodeCount: 6
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
  scalingInput:
    scaledUpNodepools:
      - quantity: 3
        etcd: true
        controlplane: false
        worker: false
      - quantity: 2
        etcd: false
        controlplane: true
        worker: false
      - quantity: 3
        etcd: false
        controlplane: false
        worker: true
    scaledDownNodepools:
      - quantity: 3
        etcd: true
        controlplane: false
        worker: false
      - quantity: 2
        etcd: false
        controlplane: true
        worker: false
      - quantity: 1
        etcd: false
        controlplane: false
        worker: true
```
Note: In this test suite, Terraform explicitly cleans up resources after each test case is performed. This is because Terraform will experience caching issues, causing tests to fail.

---

<a name="configurations-terratest-kubernetes_upgrade"></a>
#### :small_red_triangle: [Back to top](#top)

##### Kubernetes Upgrade

```yaml
terratest:
  pathToRepo: # REQUIRED - path to repo from user's go directory i.e. ../go/<path/to/repo/tfp-automation>
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
  nodeCount: 3
  kubernetesVersion: ""
  upgradedKubernetesVersion: ""
```
Note: In this test suite, Terraform explicitly cleans up resources after each test case is performed. This is because Terraform will experience caching issues, causing tests to fail.

---

<a name="configurations-terratest-snapshots"></a>
#### :small_red_triangle: [Back to top](#top)

##### ETCD Snapshots

```yaml
terratest:
  pathToRepo: # REQUIRED - path to repo from user's go directory i.e. ../go/<path/to/repo/tfp-automation>
  snapshotInput:
    snapshotRestore: "none"
    upgradeKubernetesVersion: ""
    controlPlaneConcurrencyValue: "15%"
    workerConcurrencyValue: "20%"
```
Note: In this test suite, Terraform explicitly cleans up resources after each test case is performed. This is because Terraform will experience caching issues, causing tests to fail.

---

<a name="configurations-terratest-build_module"></a>
#### :small_red_triangle: [Back to top](#top)

##### Build Module

Build module test may be used and ran to create a main.tf terraform configuration file for the desired module.  This is logged to the output for future reference and use.

Testing configurations for this are the same as outlined in provisioning test above.  Please review provisioning test configurations for more details.

---

<a name="configurations-terratest-cleanup"></a>
#### :small_red_triangle: [Back to top](#top)

##### Cleanup

Cleanup test may be used to clean up resources in situations where rancher config has `cleanup` set to `false`.  This may be helpful in debugging. This test expects the same configurations used to initially create this environment, to properly clean them up.