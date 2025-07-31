package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetVsphereRKE1Provider is a helper function that will set the Vsphere RKE1
// Terraform provider details in the main.tf file.
func SetVsphereRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	vsphereConfigBlock := nodeTemplateBlockBody.AppendNewBlock(vsphere.VsphereConfig, nil)
	vsphereConfigBlockBody := vsphereConfigBlock.Body()

	cfgparams := make([]cty.Value, len(terraformConfig.VsphereConfig.Cfgparam))
	for i, cfgparam := range terraformConfig.VsphereConfig.Cfgparam {
		cfgparams[i] = cty.StringVal(cfgparam)
	}

	networks := make([]cty.Value, len(terraformConfig.VsphereConfig.Network))
	for i, network := range terraformConfig.VsphereConfig.Network {
		networks[i] = cty.StringVal(network)
	}

	vsphereConfigBlockBody.SetAttributeValue(vsphere.DockerURL, cty.StringVal(terraformConfig.VsphereConfig.Boot2dockerURL))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Cfgparam, cty.ListVal(cfgparams))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.CloneFrom, cty.StringVal(terraformConfig.VsphereConfig.CloneFrom))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.CloudConfig, cty.StringVal(terraformConfig.VsphereConfig.CloudConfig))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Cloudinit, cty.StringVal(terraformConfig.VsphereConfig.Cloudinit))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.ContentLibrary, cty.StringVal(terraformConfig.VsphereConfig.ContentLibrary))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.CPUCount, cty.StringVal(terraformConfig.VsphereConfig.CPUCount))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.CreationType, cty.StringVal(terraformConfig.VsphereConfig.CreationType))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.DataCenter, cty.StringVal(terraformConfig.VsphereConfig.DataCenter))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.DataStore, cty.StringVal(terraformConfig.VsphereConfig.DataStore))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.DatastoreCluster, cty.StringVal(terraformConfig.VsphereConfig.DatastoreCluster))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.DiskSize, cty.StringVal(terraformConfig.VsphereConfig.DiskSize))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Folder, cty.StringVal(terraformConfig.VsphereConfig.Folder))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.HostSystem, cty.StringVal(terraformConfig.VsphereConfig.HostSystem))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.MemorySize, cty.StringVal(terraformConfig.VsphereConfig.MemorySize))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Network, cty.ListVal(networks))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Password, cty.StringVal(terraformConfig.VsphereCredentials.Password))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Pool, cty.StringVal(terraformConfig.VsphereConfig.Pool))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.SSHPassword, cty.StringVal(terraformConfig.VsphereConfig.SSHPassword))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.SSHPort, cty.StringVal(terraformConfig.VsphereConfig.SSHPort))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.SSHUser, cty.StringVal(terraformConfig.VsphereConfig.SSHUser))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.SSHUserGroup, cty.StringVal(terraformConfig.VsphereConfig.SSHUserGroup))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Username, cty.StringVal(terraformConfig.VsphereCredentials.Username))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.Vcenter, cty.StringVal(terraformConfig.VsphereCredentials.Vcenter))
	vsphereConfigBlockBody.SetAttributeValue(vsphere.VcenterPort, cty.StringVal(terraformConfig.VsphereCredentials.VcenterPort))
}

// SetVsphereRKE2K3SProvider is a helper function that will set the Vsphere RKE2/K3S Terraform provider details in the main.tf file.
func SetVsphereRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.StandardUser)},
	}

	cloudCredBlockBody.SetAttributeRaw(defaults.Provider, provider)
	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	vsphereCredBlock := cloudCredBlockBody.AppendNewBlock(vsphere.VsphereCredentialConfig, nil)
	vsphereCredBlockBody := vsphereCredBlock.Body()

	vsphereCredBlockBody.SetAttributeValue(vsphere.Password, cty.StringVal(terraformConfig.VsphereCredentials.Password))
	vsphereCredBlockBody.SetAttributeValue(vsphere.Username, cty.StringVal(terraformConfig.VsphereCredentials.Username))
	vsphereCredBlockBody.SetAttributeValue(vsphere.Vcenter, cty.StringVal(terraformConfig.VsphereCredentials.Vcenter))
	vsphereCredBlockBody.SetAttributeValue(vsphere.VcenterPort, cty.StringVal(terraformConfig.VsphereCredentials.VcenterPort))
}
