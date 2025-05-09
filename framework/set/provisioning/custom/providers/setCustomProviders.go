package providers

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetCustomProviders is a helper function that will set the general Terraform provider configurations in the main.tf file.
func SetCustomProviders(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")
	localProviderVersion := os.Getenv("LOCALS_PROVIDER_VERSION")
	rancher2ProviderVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")

	rancher2Source := defaults.Rancher2Source
	if strings.Contains(rancher2ProviderVersion, defaults.Rc) {
		rancher2Source = defaults.Rancher2LocalSource
	}

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(defaults.Terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(defaults.RequiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.AwsSource),
		defaults.Version: cty.StringVal(cloudProviderVersion),
	}))

	reqProvsBlockBody.SetAttributeValue(defaults.Local, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.LocalSource),
		defaults.Version: cty.StringVal(localProviderVersion),
	}))

	reqProvsBlockBody.SetAttributeValue(defaults.Rancher2, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(rancher2Source),
		defaults.Version: cty.StringVal(rancher2ProviderVersion),
	}))

	rootBody.AppendNewline()

	awsProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Aws})
	awsProvBlockBody := awsProvBlock.Body()

	awsProvBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsProvBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsProvBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

	rootBody.AppendNewline()

	rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Local})

	rootBody.AppendNewline()

	rancher2ProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Rancher2})
	rancher2ProvBlockBody := rancher2ProvBlock.Body()

	rancher2ProvBlockBody.SetAttributeValue(defaults.ApiUrl, cty.StringVal(`https://`+rancherConfig.Host))
	rancher2ProvBlockBody.SetAttributeValue(defaults.TokenKey, cty.StringVal(rancherConfig.AdminToken))
	rancher2ProvBlockBody.SetAttributeValue(defaults.Insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()

	return newFile, rootBody
}
