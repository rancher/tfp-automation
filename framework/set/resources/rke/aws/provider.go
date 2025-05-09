package aws

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// createTerraformProviderBlock will up the terraform block with the required aws provider.
func createTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv(cloudProviderEnvVar)
	if cloudProviderVersion == "" {
		logrus.Fatalf("Expected env var not set %s", cloudProviderEnvVar)
	}

	rkeProviderVersion := os.Getenv(rkeProviderEnvVar)
	if rkeProviderVersion == "" {
		logrus.Fatalf("Expected env var not set %s", rkeProviderEnvVar)
	}

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.AwsSource),
		defaults.Version: cty.StringVal(cloudProviderVersion),
	}))

	source := "rancher/rke"
	if strings.Contains(rkeProviderVersion, rc) {
		source = "terraform.local/local/rke"
	}

	reqProvsBlockBody.SetAttributeValue(defaults.RKE, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(source),
		defaults.Version: cty.StringVal(rkeProviderVersion),
	}))
}

// createRKEProviderBlock will set up the RKE1 provider block.
func createRKEProviderBlock(rootBody *hclwrite.Body) {
	rkeProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.RKE})
	rkeProvBlockBody := rkeProvBlock.Body()

	rkeProvBlockBody.SetAttributeValue(defaults.LogFile, cty.StringVal(rkeLogFile))
}
