# <p align="center">RANCHER :small_blue_diamond: TFP-AUTOMATION</p>

`tfp-automation` is a framework designed to test various Rancher2 Terraform provider resources to be tested with Terratest + Go. While this is not meant to serve as a 1:1 partiy with the existing test cases in `rancher/rancher`, the overall structure of the tests is. This is to ensure that adoption of the framework is as seamless as possible.

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

###### The Rancher2 provider version is determined via the `RANCHER2_PROVIDER_VERSION` environment variable.

##### When testing locally, it is required to set the `RANCHER2_PROVIDER_VERSION`, as type `string`, and formatted without a leading `v`.

##### Example: `export RANCHER2_PROVIDER_VERSION="4.2.0"` or `export RANCHER2_PROVIDER_VERSION="4.2.0-rc3"`

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

The `rancher` configurations in the `cattle-config.yaml` will remain consistent across all modules and tests.  Fields to configure in this section are as follows:

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>host</td>
      <td>url to rancher sercer without leading https:// and without trailing /</td>
      <td>string</td>
      <td>url-to-rancher-server.com</td>
    </tr>
    <tr>
      <td>adminToken</td>
      <td>rancher admin bearer token</td>
      <td>string</td>
      <td>token-XXXXX:XXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>insecure</td>
      <td>must be set to true</td>
      <td>boolean</td>
      <td>true</td>
    </tr>
    <tr>
      <td>cleanup</td>
      <td>If true, resources will be cleaned up upon test completion</td>
      <td>boolean</td>
      <td>true</td>
    </tr>
  </tbody>
</table>

##### Example:

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

##### Example:

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

Module specific fields to configure in the terraform section are as follows:

<a name="configurations-terraform-aks"></a>
#### :small_red_triangle: [Back to top](#top)

##### AKS

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>aks</td>
    </tr>
    <tr>
      <td>cloudCredentialName</td>
      <td>provide the name of unique cloud credentials to be created during testing</td>
      <td>string</td>
      <td>tf-aks</td>
    </tr>
    <tr>
      <td>clientID</td>
      <td>provide azure client id</td>
      <td>string</td>
      <td>XXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>clientSecret</td>
      <td>provide azure client secret</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>subscriptionID</td>
      <td>provide azure subscription id</td>
      <td>string</td>
      <td>XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>resourceGroup</td>
      <td>provide an existing resource group from Azure</td>
      <td>string</td>
      <td>my-resource-group</td>
    </tr>
    <tr>
      <td>resourceLocation</td>
      <td>provide location for Azure instances</td>
      <td>string</td>
      <td>eastus</td>
    </tr>
    <tr>
      <td>hostnamePrefix</td>
      <td>provide a unique hostname prefix for resources</td>
      <td>string</td>
      <td>tfp</td>
    </tr>
    <tr>
      <td>networkPlugin</td>
      <td>provide network plugin</td>
      <td>string</td>
      <td>kubenet</td>
    </tr>
    <tr>
      <td>availabilityZones</td>
      <td>list of availablilty zones</td>
      <td>[]string</td>
      <td>
      - '1' <br/>
      - '2' <br/>
      - '3'
      </td>
    </tr>
    <tr>
      <td>osDiskSizeGB</td>
      <td>os disk size in gigabytes</td>
      <td>int64</td>
      <td>128</td>
    </tr>
    <tr>
      <td>vmSize</td>
      <td>vm size to be used for instances</td>
      <td>string</td>
      <td>Standard_DS2_v2</td>
    </tr>
  </tbody>
</table>

##### Example:

```yaml
terraform:
  module: aks
  cloudCredentialName: tf-aks
  azureConfig:
    clientID: ""
    clientSecret: ""
    subscriptionID: ""
    resourceGroup: ""
    resourceLocation: eastus
    availabilityZones:
      - '1'
      - '2'
      - '3'
    osDiskSizeGB: 128
    tenantId: ""
    vmSize: Standard_DS2_v2
```

---

<a name="configurations-terraform-eks"></a>
#### :small_red_triangle: [Back to top](#top)

##### EKS

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>eks</td>
    </tr>
    <tr>
      <td>cloudCredentialName</td>
      <td>provide the name of unique cloud credentials to be created during testing</td>
      <td>string</td>
      <td>tf-eks</td>
    </tr>
    <tr>
      <td>awsAccessKey</td>
      <td>provide aws access key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>awsSecretKey</td>
      <td>provide aws secret key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>awsInstanceType</td>
      <td>provide aws instance type</td>
      <td>string</td>
      <td>t3.medium</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide a region for resources to be created in</td>
      <td>string</td>
      <td>us-east-2</td>
    </tr>
    <tr>
      <td>awsSubnets</td>
      <td>list of valid subnet IDs</td>
      <td>[]string</td>
      <td>
        - subnet-xxxxxxxx <br/>
        - subnet-yyyyyyyy <br/>
        - subnet-zzzzzzzz
      </td>
    </tr>
    <tr>
      <td>awsSecurityGroups</td>
      <td>list of security group IDs to be applied to AWS instances</td>
      <td>[]string</td>
      <td>- sg-xxxxxxxxxxxxxxxxx</td>
    </tr>
    <tr>
      <td>hostnamePrefix</td>
      <td>provide a unique hostname prefix for resources</td>
      <td>string</td>
      <td>tfp</td>
    </tr>
    <tr>
      <td>publicAccess</td>
      <td>If true, public access will be enabled</td>
      <td>boolean</td>
      <td>true</td>
    </tr>
    <tr>
      <td>privateAccess</td>
      <td>If true, private access will be enabled</td>
      <td>boolean</td>
      <td>true</td>
    </tr>
    <tr>
      <td>nodeRole</td>
      <td>Optional with Rancher v2.7+ - if provided, this custom role will be used when creating instances for node groups</td>
      <td>string</td>
      <td>arn:aws:iam::############:role/my-custom-NodeInstanceRole-############</td>
    </tr>
  </tbody>
</table>

##### Example:

```yaml
terraform:
  module: eks
  cloudCredentialName: tf-eks
  hostnamePrefix: tfp
  awsConfig:
    awsAccessKey: ""
    awsSecretKey: ""
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

##### GKE


<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>gke</td>
    </tr>
    <tr>
      <td>cloudCredentialName</td>
      <td>provide the name of unique cloud credentials to be created during testing</td>
      <td>string</td>
      <td>tf-gke</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide region for resources to be created in</td>
      <td>string</td>
      <td>us-central1-c</td>
    </tr>
    <tr>
      <td>projectID</td>
      <td>provide gke project ID</td>
      <td>string</td>
      <td>my-project-id-here</td>
    </tr>
    <tr>
      <td>network</td>
      <td>specify network here</td>
      <td>string</td>
      <td>default</td>
    </tr>
    <tr>
      <td>subnetwork</td>
      <td>specify subnetwork here</td>
      <td>string</td>
      <td>default</td>
    </tr>
    <tr>
      <td>hostnamePrefix</td>
      <td>provide a unique hostname prefix for resources</td>
      <td>string</td>
      <td>tfp</td>
    </tr>
  </tbody>
</table>

##### Example:

```yaml
terraform:
  module: gke
  cloudCredentialName: tf-creds-gke
  hostnamePrefix: tfp
  googleConfig:
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
    region: us-central1-c
    projectID: ""
    network: default
    subnetwork: default
```

---

<a name="configurations-terraform-azure_rke1"></a>
#### :small_red_triangle: [Back to top](#top)

##### AZURE_RKE1

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke1</td>
    </tr>
    <tr>
      <td>availabilitySet</td>
      <td>provide availability set to put virtual machine in</td>
      <td>string</td>
      <td>docker-machine</td>
    </tr>
    <tr>
      <td>clientId</td>
      <td>provide client ID</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>clientSecret</td>
      <td>provide client secret</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>subscriptionId</td>
      <td>provide subscription ID</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>customData</td>
      <td>provide path to file</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>diskSize</td>
      <td>disk size if using managed disk</td>
      <td>string</td>
      <td>100</td>
    </tr>
    <tr>
      <td>dockerPort</td>
      <td>port number for Docker engine</td>
      <td>string</td>
      <td>2376</td>
    </tr>
    <tr>
      <td>environment</td>
      <td>Azure environment</td>
      <td>string</td>
      <td>AzurePublicCloud</td>
    </tr>
    <tr>
      <td>faultDomainCount</td>
      <td>fault domain count to use for availability set</td>
      <td>string</td>
      <td>3</td>
    </tr>
    <tr>
      <td>image</td>
      <td>Azure virtual machine OS image</td>
      <td>string</td>
      <td>Canonical:0001-com-ubuntu-server-jammy:22_04-lts:latest</td>
    </tr>
    <tr>
      <td>location</td>
      <td>Azure region to create virtual machines</td>
      <td>string</td>
      <td>eastus2</td>
    </tr>
    <tr>
      <td>managedDisks</td>
      <td>configures VM and availability set for managed disks</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>noPublicIp</td>
      <td>do not create a public IP address for the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>openPort</td>
      <td>make the specified port number accessible from the Internet</td>
      <td>list</td>
      <td>false</td>
    </tr>
    <tr>
      <td>privateIpAddress</td>
      <td>specify a static private IP address for the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>resourceGroup</td>
      <td>provide a Azure resource group</td>
      <td>string</td>
      <td>docker-machine</td>
    </tr>
    <tr>
      <td>size</td>
      <td>size for Azure virtual machine</td>
      <td>string</td>
      <td>Standard_A2</td>
    </tr>
    <tr>
      <td>sshUser</td>
      <td>ssh username</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>staticPublicIp</td>
      <td>Assign a static public IP address to the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>storageType</td>
      <td>type of Storage Account to host the OS Disk for the machine</td>
      <td>string</td>
      <td>Standard_LRS</td>
    </tr>
    <tr>
      <td>updateDomainCount</td>
      <td>update domain count to use for availability set</td>
      <td>string</td>
      <td>3</td>
    </tr>
  </tbody>
</table>

##### Example:

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

##### EC2_RKE1

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke1</td>
    </tr>
    <tr>
      <td>awsAccessKey</td>
      <td>provide aws access key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>awsSecretKey</td>
      <td>provide aws secret key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>ami</td>
      <td>provide ami; (optional - may be left as empty string '')</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>awsInstanceType</td>
      <td>provide aws instance type</td>
      <td>string</td>
      <td>t3.medium</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide a region for resources to be created in</td>
      <td>string</td>
      <td>us-east-2</td>
    </tr>
    <tr>
      <td>awsSecurityGroupNames</td>
      <td>list of security groups to be applied to AWS instances</td>
      <td>[]string</td>
      <td>- security-group-name</td>
    </tr>
    <tr>
      <td>awsSubnetID</td>
      <td>provide a valid subnet ID</td>
      <td>string</td>
      <td>subnet-xxxxxxxx</td>
    </tr>
    <tr>
      <td>awsVpcID</td>
      <td>provide a valid VPC ID</td>
      <td>string</td>
      <td>vpc-xxxxxxxx</td>
    </tr>
    <tr>
      <td>awsZoneLetter</td>
      <td>provide zone letter to be used</td>
      <td>string</td>
      <td>a</td>
    </tr>
    <tr>
      <td>awsRootSize</td>
      <td>root size in gigabytes</td>
      <td>int64</td>
      <td>80</td>
    </tr>
    <tr>
      <td>networkPlugin</td>
      <td>provide network plugin to be used</td>
      <td>string</td>
      <td>canal</td>
    </tr>
    <tr>
      <td>nodeTemplateName</td>
      <td>provide a unique name for node template</td>
      <td>string</td>
      <td>tf-rke1-template</td>
    </tr>
    <tr>
      <td>hostnamePrefix</td>
      <td>provide a unique hostname prefix for resources</td>
      <td>string</td>
      <td>tfp</td>
    </tr>
  </tbody>
</table>

##### Example:

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

##### LINODE_RKE1

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>linode_rke1</td>
    </tr>
   <tr>
      <td>linodeToken</td>
      <td>provide linode token credential</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide a region for resources to be created in</td>
      <td>string</td>
      <td>us-east</td>
    </tr>
    <tr>
      <td>linodeRootPass</td>
      <td>provide a unique root password</td>
      <td>string</td>
      <td>xxxxxxxxxxxxxxxx</td>
    </tr>
    <tr>
      <td>networkPlugin</td>
      <td>provide network plugin to be used</td>
      <td>string</td>
      <td>canal</td>
    </tr>
    <tr>
      <td>nodeTemplateName</td>
      <td>provide a unique name for node template</td>
      <td>string</td>
      <td>tf-rke1-template</td>
    </tr>
    <tr>
      <td>hostnamePrefix</td>
      <td>provide a unique hostname prefix for resources</td>
      <td>string</td>
      <td>tfp</td>
    </tr>
  </tbody>
</table>

##### Example:
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

##### VSPHERE_RKE1

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke1</td>
    </tr>
    <tr>
      <td>cfgparam</td>
      <td>vSphere vm configuration parameters</td>
      <td>list</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cloneFrom</td>
      <td>name of what VM you want to clone</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cloudConfig</td>
      <td>cloud config YAML content to inject as user-data</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>contentLibrary</td>
      <td>specify the name of the library</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cpuCount</td>
      <td>vSphere CPU number for docker VM</td>
      <td>string</td>
      <td>2</td>
    </tr>
    <tr>
      <td>creationType</td>
      <td>disk size if using managed disk</td>
      <td>string</td>
      <td>100</td>
    </tr>
    <tr>
      <td>datacenter</td>
      <td>vSphere datacenter for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>datastore</td>
      <td>vSphere datastore for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>datastoreCluster</td>
      <td>fvSphere datastore cluster for VM</td>
      <td>string</td>
      <td>3</td>
    </tr>
    <tr>
      <td>diskSize</td>
      <td>vSphere size of disk for docker VM (in MB)</td>
      <td>string</td>
      <td>2048</td>
    </tr>
    <tr>
      <td>folder</td>
      <td>vSphere folder for the docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>hostsystem</td>
      <td>vSphere compute resource where the docker VM will be instantiated</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>memorySize</td>
      <td>vSphere size of memory for docker VM (in MB)</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>network</td>
      <td>vSphere network where the docker VM will be attached</td>
      <td>list</td>
      <td>''</td>
    </tr>
    <tr>
      <td>password</td>
      <td>specify the vSphere password</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>pool</td>
      <td>vSphere resource pool for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>sshPassword</td>
      <td>specify the ssh password</td>
      <td>string</td>
      <td>tc_user</td>
    </tr>
    <tr>
      <td>sshPort</td>
      <td>specify the ssh port</td>
      <td>string</td>
      <td>22</td>
    </tr>
    <tr>
      <td>sshUser</td>
      <td>specify the ssh user</td>
      <td>string</td>
      <td>docker</td>
    </tr>
    <tr>
      <td>sshUserGroup</td>
      <td>specify the ssh user group</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>username</td>
      <td>specify the vSphere username</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>vcenter</td>
      <td>specify the vcenter</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>vcenterPort</td>
      <td>specify the vcenter port</td>
      <td>string</td>
      <td>44</td>
    </tr>
  </tbody>
</table>

##### Example:

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

##### AZURE_RKE2 + AZURE_K3S

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke1</td>
    </tr>
    <tr>
      <td>availabilitySet</td>
      <td>provide availability set to put virtual machine in</td>
      <td>string</td>
      <td>docker-machine</td>
    </tr>
    <tr>
      <td>clientId</td>
      <td>provide client ID</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>clientSecret</td>
      <td>provide client secret</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>subscriptionId</td>
      <td>provide subscription ID</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>customData</td>
      <td>provide path to file</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>diskSize</td>
      <td>disk size if using managed disk</td>
      <td>string</td>
      <td>100</td>
    </tr>
    <tr>
      <td>dockerPort</td>
      <td>port number for Docker engine</td>
      <td>string</td>
      <td>2376</td>
    </tr>
    <tr>
      <td>environment</td>
      <td>Azure environment</td>
      <td>string</td>
      <td>AzurePublicCloud</td>
    </tr>
    <tr>
      <td>faultDomainCount</td>
      <td>fault domain count to use for availability set</td>
      <td>string</td>
      <td>3</td>
    </tr>
    <tr>
      <td>image</td>
      <td>Azure virtual machine OS image</td>
      <td>string</td>
      <td>Canonical:0001-com-ubuntu-server-jammy:22_04-lts:latest</td>
    </tr>
    <tr>
      <td>location</td>
      <td>Azure region to create virtual machines</td>
      <td>string</td>
      <td>eastus2</td>
    </tr>
    <tr>
      <td>managedDisks</td>
      <td>configures VM and availability set for managed disks</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>noPublicIp</td>
      <td>do not create a public IP address for the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>openPort</td>
      <td>make the specified port number accessible from the Internet</td>
      <td>list</td>
      <td>false</td>
    </tr>
    <tr>
      <td>privateIpAddress</td>
      <td>specify a static private IP address for the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>resourceGroup</td>
      <td>provide a Azure resource group</td>
      <td>string</td>
      <td>docker-machine</td>
    </tr>
    <tr>
      <td>size</td>
      <td>size for Azure virtual machine</td>
      <td>string</td>
      <td>Standard_A2</td>
    </tr>
    <tr>
      <td>sshUser</td>
      <td>ssh username</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>staticPublicIp</td>
      <td>Assign a static public IP address to the machine</td>
      <td>bool</td>
      <td>false</td>
    </tr>
    <tr>
      <td>storageType</td>
      <td>type of Storage Account to host the OS Disk for the machine</td>
      <td>string</td>
      <td>Standard_LRS</td>
    </tr>
    <tr>
      <td>tenantId</td>
      <td>provide the tenant ID</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>updateDomainCount</td>
      <td>update domain count to use for availability set</td>
      <td>string</td>
      <td>3</td>
    </tr>
  </tbody>
</table>

##### Example:

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

##### EC2_RKE2 + EC2_K3S

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke2</td>
    </tr>
    <tr>
      <td>cloudCredentialName</td>
      <td>provide the name of unique cloud credentials to be created during testing</td>
      <td>string</td>
      <td>tf-creds-rke2</td>
    </tr>
    <tr>
      <td>awsAccessKey</td>
      <td>provide aws access key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>awsSecretKey</td>
      <td>provide aws secret key</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>ami</td>
      <td>provide ami; (optional - may be left as empty string '')</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide a region for resources to be created in</td>
      <td>string</td>
      <td>us-east-2</td>
    </tr>
    <tr>
      <td>awsSecurityGroupNames</td>
      <td>list of security groups to be applied to AWS instances</td>
      <td>[]string</td>
      <td>- my-security-group</td>
    </tr>
    <tr>
      <td>awsSubnetID</td>
      <td>provide a valid subnet ID</td>
      <td>string</td>
      <td>subnet-xxxxxxxx</td>
    </tr>
    <tr>
      <td>awsVpcID</td>
      <td>provide a valid VPC ID</td>
      <td>string</td>
      <td>vpc-xxxxxxxx</td>
    </tr>
    <tr>
      <td>awsZoneLetter</td>
      <td>provide zone letter to be used</td>
      <td>string</td>
      <td>a</td>
    </tr>
    <tr>
      <td>machineConfigName</td>
      <td>provide a unique name for machine config</td>
      <td>string</td>
      <td>tf-rke2</td>
    </tr>
    <tr>
      <td>enableNetworkPolicy</td>
      <td>If true, Network Policy will be enabled</td>
      <td>boolean</td>
      <td>false</td>
    </tr>
    <tr>
      <td>defaultClusterRoleForProjectMembers</td>
      <td>select default role to be used for project memebers</td>
      <td>string</td>
      <td>user</td>
    </tr>
  </tbody>
</table>

##### Example:
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

##### LINODE_RKE2 + LINODE_K3S

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>linode_k3s</td>
    </tr>
    <tr>
      <td>cloudCredentialName</td>
      <td>provide the name of unique cloud credentials to be created during testing</td>
      <td>string</td>
      <td>tf-linode</td>
    </tr>
    <tr>
      <td>linodeToken</td>
      <td>provide linode token credential</td>
      <td>string</td>
      <td>XXXXXXXXXXXXXXXXXXXX</td>
    </tr>
    <tr>
      <td>linodeImage</td>
      <td>specify image to be used for instances</td>
      <td>string</td>
      <td>linode/ubuntu20.04</td>
    </tr>
    <tr>
      <td>region</td>
      <td>provide a region for resources to be created in</td>
      <td>string</td>
      <td>us-east</td>
    </tr>
    <tr>
      <td>linodeRootPass</td>
      <td>provide a unique root password</td>
      <td>string</td>
      <td>xxxxxxxxxxxxxxxx</td>
    </tr>
    <tr>
      <td>machineConfigName</td>
      <td>provide a unique name for machine config</td>
      <td>string</td>
      <td>tf-k3s</td>
    </tr>
    <tr>
      <td>enableNetworkPolicy</td>
      <td>If true, Network Policy will be enabled</td>
      <td>boolean</td>
      <td>false</td>
    </tr>
    <tr>
      <td>defaultClusterRoleForProjectMembers</td>
      <td>select default role to be used for project memebers</td>
      <td>string</td>
      <td>user</td>
    </tr>
  </tbody>
</table>

##### Example:
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

##### VSPHERE_RKE2 + VSPHERE_K3S

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>module</td>
      <td>specify terraform module here</td>
      <td>string</td>
      <td>ec2_rke1</td>
    </tr>
    <tr>
      <td>cfgparam</td>
      <td>vSphere vm configuration parameters</td>
      <td>list</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cloneFrom</td>
      <td>name of what VM you want to clone</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cloudConfig</td>
      <td>cloud config YAML content to inject as user-data</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>contentLibrary</td>
      <td>specify the name of the library</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>cpuCount</td>
      <td>vSphere CPU number for docker VM</td>
      <td>string</td>
      <td>2</td>
    </tr>
    <tr>
      <td>creationType</td>
      <td>disk size if using managed disk</td>
      <td>string</td>
      <td>100</td>
    </tr>
    <tr>
      <td>datacenter</td>
      <td>vSphere datacenter for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>datastore</td>
      <td>vSphere datastore for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>datastoreCluster</td>
      <td>fvSphere datastore cluster for VM</td>
      <td>string</td>
      <td>3</td>
    </tr>
    <tr>
      <td>diskSize</td>
      <td>vSphere size of disk for docker VM (in MB)</td>
      <td>string</td>
      <td>2048</td>
    </tr>
    <tr>
      <td>folder</td>
      <td>vSphere folder for the docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>hostsystem</td>
      <td>vSphere compute resource where the docker VM will be instantiated</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>memorySize</td>
      <td>vSphere size of memory for docker VM (in MB)</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>network</td>
      <td>vSphere network where the docker VM will be attached</td>
      <td>list</td>
      <td>''</td>
    </tr>
    <tr>
      <td>password</td>
      <td>specify the vSphere password</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>pool</td>
      <td>vSphere resource pool for docker VM</td>
      <td>string</td>
      <td>''</td>
    </tr>
    <tr>
      <td>sshPassword</td>
      <td>specify the ssh password</td>
      <td>string</td>
      <td>tc_user</td>
    </tr>
    <tr>
      <td>sshPort</td>
      <td>specify the ssh port</td>
      <td>string</td>
      <td>22</td>
    </tr>
    <tr>
      <td>sshUser</td>
      <td>specify the ssh user</td>
      <td>string</td>
      <td>docker</td>
    </tr>
    <tr>
      <td>sshUserGroup</td>
      <td>specify the ssh user group</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>username</td>
      <td>specify the vSphere username</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>vcenter</td>
      <td>specify the vcenter</td>
      <td>string</td>
      <td>staff</td>
    </tr>
    <tr>
      <td>vcenterPort</td>
      <td>specify the vcenter port</td>
      <td>string</td>
      <td>44</td>
    </tr>
  </tbody>
</table>

##### Example:

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

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>nodepools</td>
      <td>provide nodepool configs to be initially provisioned</td>
      <td>[]Nodepool</td>
      <td>view section on nodepools above or example yaml below</td>
    </tr>
    <tr>
      <td>kubernetesVersion</td>
      <td>specify the kubernetes version to be used</td>
      <td>string</td>
      <td>view yaml below for all module specific expected k8s version formats</td>
    </tr>
    <tr>
      <td>nodeCount</td>
      <td>provide the expected initial node count</td>
      <td>int64</td>
      <td>3</td>
    </tr>
  </tbody>
</table>

###### Example:
```yaml
# this example is valid for RKE1 provision
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

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>nodepools</td>
      <td>provide nodepool configs to be initially provisioned</td>
      <td>[]Nodepool</td>
      <td>view section on nodepools above or example yaml below</td>
    </tr>
    <tr>
      <td>scaledUpNodepools</td>
      <td>provide nodepool configs to be scaled up to, after initial provisioning</td>
      <td>[]Nodepool</td>
      <td>view section on nodepools above or example yaml below</td>
    </tr>
    <tr>
      <td>scaledDownNodepools</td>
      <td>provide nodepool configs to be scaled down to, after scaling up cluster</td>
      <td>[]Nodepool</td>
      <td>view section on nodepools above or example yaml below</td>
    </tr>
    <tr>
      <td>kubernetesVersion</td>
      <td>specify the kubernetes version to be used</td>
      <td>string</td>
      <td>view example yaml above for provisioning test for all module specific expected k8s version formats</td>
    </tr>
    <tr>
      <td>nodeCount</td>
      <td>provide the expected initial node count</td>
      <td>int64</td>
      <td>3</td>
    </tr>
    <tr>
      <td>scaledUpNodeCount</td>
      <td>provide the expected node count of scaled up cluster</td>
      <td>int64</td>
      <td>8</td>
    </tr>
    <tr>
      <td>scaledDownNodeCount</td>
      <td>provide the expected node count of scaled down cluster</td>
      <td>int64</td>
      <td>6</td>
    </tr>
  </tbody>
</table>

###### Example:
```yaml
# this example is valid for RKE1 scale
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

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>nodepools</td>
      <td>provide nodepool configs to be initially provisioned</td>
      <td>[]Nodepool</td>
      <td>view section on nodepools above or example yaml below</td>
    </tr>
    <tr>
      <td>nodeCount</td>
      <td>provide the expected initial node count</td>
      <td>int64</td>
      <td>3</td>
    </tr>
    <tr>
      <td>kubernetesVersion</td>
      <td>specify the kubernetes version to be used</td>
      <td>string</td>
      <td>view example yaml above for provisioning test for all module specific expected k8s version formats</td>
    </tr>
    <tr>
      <td>upgradedKubernetesVersion</td>
      <td>specify the kubernetes version to be upgraded to</td>
      <td>string</td>
      <td>view example yaml above for provisioning test for all module specific expected k8s version formats</td>
    </tr>
  </tbody>
</table>

###### Example:
```yaml
# this example is valid for K3s kubernetes upgrade
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

<table>
  <tbody>
    <tr>
      <th>Field</th>
      <th>Description</th>
      <th>Type</th>
      <th>Example</th>
    </tr>
    <tr>
      <td>snapshotInput</td>
      <td>Block in which to define snapshot parameters</td>
      <td>Snapshots</td>
      <td>view section on snapshotInput in example yaml below</td>
    </tr>
    <tr>
      <td>snapshotRestore</td>
      <td>provide the snapshot restore option (none, kubernetesVersion or all)</td>
      <td>string</td>
      <td>none</td>
    </tr>
    <tr>
      <td>upgradeKubernetesVersion</td>
      <td>specify the kubernetes version to be upgraded to</td>
      <td>string</td>
      <td>view section on snapshotInput in example yaml below</td>
    </tr>
    <tr>
      <td>controlPlaneConcurrencyValue</td>
      <td>specify the control plane concurrency value used when upgrading</td>
      <td>string</td>
      <td>view section on snapshotInput in example yaml below</td>
    </tr>
    <tr>
      <td>workerConcurrencyValue</td>
      <td>specify the worker plane concurrency value used when upgrading RKE2/K3s clusters</td>
      <td>string</td>
      <td>view section on snapshotInput in example yaml below</td>
    </tr>
  </tbody>
</table>

###### Example:
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
