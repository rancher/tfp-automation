package google

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

// CreateGoogleCloudInstances will set up the Google Cloud instances.
func CreateGoogleCloudInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		configBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(totalNodeCount))
	}

	configBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-"+hostnamePrefix))
	configBlockBody.SetAttributeValue(machineType, cty.StringVal(terraformConfig.GoogleConfig.MachineType))

	bootDiskBlock := configBlockBody.AppendNewBlock(googleDefaults.GoogleBootDisk, nil)
	bootDiskBlockBody := bootDiskBlock.Body()

	initializeParamsBlock := bootDiskBlockBody.AppendNewBlock(googleDefaults.GoogleInitializeParams, nil)
	initializeParamsBlockBody := initializeParamsBlock.Body()

	initializeParamsBlockBody.SetAttributeValue(general.Image, cty.StringVal(terraformConfig.GoogleConfig.Image))
	initializeParamsBlockBody.SetAttributeValue(general.Size, cty.NumberIntVal(terraformConfig.GoogleConfig.Size))

	networkInterfaceBlock := configBlockBody.AppendNewBlock(googleDefaults.GoogleNetworkInterface, nil)
	networkInterfaceBlockBody := networkInterfaceBlock.Body()

	networkInterfaceBlockBody.SetAttributeValue(network, cty.StringVal(terraformConfig.GoogleConfig.Network))

	accessConfigBlock := networkInterfaceBlockBody.AppendNewBlock(googleDefaults.GoogleAccessConfig, nil)
	_ = accessConfigBlock.Body()

	metadataBlock := configBlockBody.AppendNewBlock(general.Metadata+" =", nil)
	metadataBlockBody := metadataBlock.Body()

	expression := fmt.Sprintf(`"%s:${file("%s")}"`, terraformConfig.Standalone.OSUser, terraformConfig.GoogleConfig.KeyPath)
	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	metadataBlockBody.SetAttributeRaw(sshKeys, value)
}
