package providers

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetCustomProviders is a helper function that will set the general Terraform provider configurations in the main.tf file.
func SetCustomProviders(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")
	localProviderVersion := os.Getenv("LOCALS_PROVIDER_VERSION")
	rancher2ProviderVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")

	rancher2Source := rancher2.Rancher2Source
	if strings.Contains(rancher2ProviderVersion, general.RC) {
		rancher2Source = rancher2.Rancher2LocalSource
	}

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(general.Terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(general.RequiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(aws.Aws, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(aws.AwsSource),
		general.Version: cty.StringVal(cloudProviderVersion),
	}))

	reqProvsBlockBody.SetAttributeValue(general.Local, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(general.LocalSource),
		general.Version: cty.StringVal(localProviderVersion),
	}))

	reqProvsBlockBody.SetAttributeValue(general.Rancher2, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(rancher2Source),
		general.Version: cty.StringVal(rancher2ProviderVersion),
	}))

	rootBody.AppendNewline()

	awsProvBlock := rootBody.AppendNewBlock(general.Provider, []string{aws.Aws})
	awsProvBlockBody := awsProvBlock.Body()

	awsProvBlockBody.SetAttributeValue(aws.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsProvBlockBody.SetAttributeValue(aws.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsProvBlockBody.SetAttributeValue(aws.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

	rootBody.AppendNewline()

	rootBody.AppendNewBlock(general.Provider, []string{general.Local})
	rootBody.AppendNewline()

	rancher2ProvBlock := rootBody.AppendNewBlock(general.Provider, []string{general.Rancher2})
	rancher2ProvBlockBody := rancher2ProvBlock.Body()

	rancher2ProvBlockBody.SetAttributeValue(general.ApiUrl, cty.StringVal(`https://`+rancherConfig.Host))
	rancher2ProvBlockBody.SetAttributeValue(general.TokenKey, cty.StringVal(rancherConfig.AdminToken))
	rancher2ProvBlockBody.SetAttributeValue(general.Insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()

	return newFile, rootBody
}
