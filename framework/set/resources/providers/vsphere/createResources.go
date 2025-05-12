package vsphere

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
)

const (
	adapterType           = "adapter_type"
	clientDevice          = "client_device"
	clone                 = "clone"
	cluster               = "cluster"
	datacenterID          = "datacenter_id"
	datastoreID           = "datastore_id"
	diskEnableUUID        = "disk_enable_uuid"
	domain                = "domain"
	guestID               = "guest_id"
	hostName              = "host_name"
	network               = "network"
	networkInterfaceTypes = "network_interface_types"
	resourcePoolID        = "resource_pool_id"
	template              = "template"
	templateUUID          = "template_uuid"
)

// CreateVsphereResources is a helper function that will create the vSphere resources needed for the RKE2 cluster.
func CreateVsphereResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateVsphereTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateVsphereProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	dataCenterExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereDatacenter + `.` + defaults.VsphereDatacenter + `.id`)
	dataCenterValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dataCenterExpression)},
	}

	CreateVsphereDatacenter(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateVsphereDatastore(rootBody, terraformConfig, dataCenterValue)
	rootBody.AppendNewline()

	CreateVsphereComputeCluster(rootBody, terraformConfig, dataCenterValue)
	rootBody.AppendNewline()

	CreateVsphereNetwork(rootBody, terraformConfig, dataCenterValue)
	rootBody.AppendNewline()

	CreateVsphereVirtualMachineTemplate(rootBody, terraformConfig, dataCenterValue)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	CreateVsphereLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}
