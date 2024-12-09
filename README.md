# <p align="center">RANCHER :small_blue_diamond: TFP-AUTOMATION</p>

`tfp-automation` is a framework designed to test various Rancher2 Terraform provider resources to be tested with Terratest + Go. While this is not meant to serve as a 1:1 partiy with the existing test cases in `rancher/rancher`, the overall structure of the tests is. This is to ensure that adoption of the framework is as seamless as possible.

In addition to the main purpose of testing the Rancher2 provider, `tfp-automation` also supports testing the RKE Terraform provider and supports AWS infrastructure creation. The latter allows for functionality such as automated daily sanity testing of the Rancher2 provider.

---

<a name="top"></a>

# <p align="center"> :scroll: Table of contents </p>

-   [Configurations](#configurations)
    -   [Rancher](#configurations-rancher)
    -   [Terraform](#configurations-terraform)
        -   [AKS](#configurations-terraform-aks)
        -   [EKS](#configurations-terraform-eks)
        -   [GKE](#configurations-terraform-gke)
        -   [AZURE_RKE1](#configurations-terraform-azure_rke1)
        -   [EC2_RKE1](#configurations-terraform-ec2_rke1)
        -   [LINODE_RKE1](#configurations-terraform-linode_rke1)
        -   [VSPHERE_RKE1](#configurations-terraform-vsphere_rke1)
        -   [AZURE_RKE2 + AZURE_K3S](#configurations-terraform-rke2_k3s_azure)
        -   [EC2_RKE2 + EC2_K3S](#configurations-terraform-rke2_k3s_ec2)
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
export RANCHER2_KEY_PATH="<tfp-automation repo path>/modules/rancher"  #Required 
export RANCHER2_PROVIDER_VERSION=""                                    #Required
export AWS_PROVIDER_VERSION=""                                         #Required for custom cluster provisioning
export LOCALS_PROVIDER_VERSION=""                                      #Required for custom cluster provisioning

export QASE_AUTOMATION_TOKEN=""                                        #Required for local Qase reporting
export QASE_TEST_RUN_ID=""                                             #Required for local Qase reporting
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
    systemDefaultRegistry: ""                 # RKE2/K3S specific
    username: ""                              # RKE1 specific
    password: ""                              # RKE1 specific
    insecure: true
    authConfigSecretName: ""                  # RKE2/K3S specific. Secret must be created in the fleet-default namespace already
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

Cleanup test may be used to clean up resources in situations where rancher config has `cleanup` set to `false`.  This may be helpful in debugging.  This test expects the same configurations used to initially create this environment, to properly clean them up.
