package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// SetVsphereRKE2K3SMachineConfig is a helper function that will set the Vsphere RKE2/K3S Terraform machine configurations in the main.tf file.
func SetVsphereRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	vsphereConfigBlock := machineConfigBlockBody.AppendNewBlock("vsphere_config", nil)
	vsphereConfigBlockBody := vsphereConfigBlock.Body()

	cfgparams := make([]cty.Value, len(terraformConfig.VsphereConfig.Cfgparam))
	for i, cfgparam := range terraformConfig.VsphereConfig.Cfgparam {
		cfgparams[i] = cty.StringVal(cfgparam)
	}

	networks := make([]cty.Value, len(terraformConfig.VsphereConfig.Network))
	for i, network := range terraformConfig.VsphereConfig.Network {
		networks[i] = cty.StringVal(network)
	}

	vsphereConfigBlockBody.SetAttributeValue("boot2docker_url", cty.StringVal(terraformConfig.VsphereConfig.Boot2dockerURL))
	vsphereConfigBlockBody.SetAttributeValue("cfgparam", cty.ListVal(cfgparams))
	vsphereConfigBlockBody.SetAttributeValue("clone_from", cty.StringVal(terraformConfig.VsphereConfig.CloneFrom))
	vsphereConfigBlockBody.SetAttributeValue("cloud_config", cty.StringVal(terraformConfig.VsphereConfig.CloudConfig))
	vsphereConfigBlockBody.SetAttributeValue("cloudinit", cty.StringVal(terraformConfig.VsphereConfig.Cloudinit))
	vsphereConfigBlockBody.SetAttributeValue("content_library", cty.StringVal(terraformConfig.VsphereConfig.ContentLibrary))
	vsphereConfigBlockBody.SetAttributeValue("cpu_count", cty.StringVal(terraformConfig.VsphereConfig.CPUCount))
	vsphereConfigBlockBody.SetAttributeValue("creation_type", cty.StringVal(terraformConfig.VsphereConfig.CreationType))
	vsphereConfigBlockBody.SetAttributeValue("datacenter", cty.StringVal(terraformConfig.VsphereConfig.DataCenter))
	vsphereConfigBlockBody.SetAttributeValue("datastore", cty.StringVal(terraformConfig.VsphereConfig.DataStore))
	vsphereConfigBlockBody.SetAttributeValue("datastore_cluster", cty.StringVal(terraformConfig.VsphereConfig.DatastoreCluster))
	vsphereConfigBlockBody.SetAttributeValue("disk_size", cty.StringVal(terraformConfig.VsphereConfig.DiskSize))
	vsphereConfigBlockBody.SetAttributeValue("folder", cty.StringVal(terraformConfig.VsphereConfig.Folder))
	vsphereConfigBlockBody.SetAttributeValue("hostsystem", cty.StringVal(terraformConfig.VsphereConfig.HostSystem))
	vsphereConfigBlockBody.SetAttributeValue("memory_size", cty.StringVal(terraformConfig.VsphereConfig.MemorySize))
	vsphereConfigBlockBody.SetAttributeValue("network", cty.ListVal(networks))
	vsphereConfigBlockBody.SetAttributeValue("password", cty.StringVal(terraformConfig.VsphereConfig.Password))
	vsphereConfigBlockBody.SetAttributeValue("pool", cty.StringVal(terraformConfig.VsphereConfig.Pool))
	vsphereConfigBlockBody.SetAttributeValue("ssh_password", cty.StringVal(terraformConfig.VsphereConfig.SSHPassword))
	vsphereConfigBlockBody.SetAttributeValue("ssh_port", cty.StringVal(terraformConfig.VsphereConfig.SSHPort))
	vsphereConfigBlockBody.SetAttributeValue("ssh_user", cty.StringVal(terraformConfig.VsphereConfig.SSHUser))
	vsphereConfigBlockBody.SetAttributeValue("ssh_user_group", cty.StringVal(terraformConfig.VsphereConfig.SSHUserGroup))
}
