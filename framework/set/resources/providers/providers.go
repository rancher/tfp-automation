package providers

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
)

type ProviderResourceFunc func(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error)

type ProviderResources struct {
	CreateAirgap    ProviderResourceFunc
	CreateNonAirgap ProviderResourceFunc
}

// TunnelToProvider returns an struct that allows a user to create resources from a given provider
func TunnelToProvider(provider string) ProviderResources {
	switch provider {
	case defaults.Aws:
		logrus.Info("Using AWS to create resources...")
		return ProviderResources{
			CreateAirgap:    aws.CreateAirgappedAWSResources,
			CreateNonAirgap: aws.CreateAWSResources,
		}
	default:
		panic(fmt.Sprintf("Provider %v not found", provider))
	}

}
