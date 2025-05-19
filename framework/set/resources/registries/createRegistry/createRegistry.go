package createRegistry

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	authRegistry    = "auth_registry"
	nonAuthRegistry = "non_auth_registry"
	ecrRegistry     = "ecr_registry"
)

// CreateAuthenticatedRegistry is a helper function that will create an authenticated registry.
func CreateAuthenticatedRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2AuthRegistryPublicDNS string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/", terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/auth-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2AuthRegistryPublicDNS, authRegistry)

	command := "bash -c '/tmp/auth-registry.sh " + terraformConfig.StandaloneRegistry.RegistryUsername + " " +
		terraformConfig.StandaloneRegistry.RegistryPassword + " " + terraformConfig.StandaloneRegistry.RegistryName + " " +
		rke2AuthRegistryPublicDNS + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.StandaloneRegistry.AssetsPath + " " + terraformConfig.Standalone.OSUser + " " +
		terraformConfig.Standalone.RancherImage

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(registryScriptContent) + "' > /tmp/auth-registry.sh"),
		cty.StringVal("chmod +x /tmp/auth-registry.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// CreateNonAuthenticatedRegistry is a helper function that will create a non-authenticated registry.
func CreateNonAuthenticatedRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2NonAuthRegistryPublicDNS, registryType string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/", terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/non-auth-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2NonAuthRegistryPublicDNS, registryType)

	var command string

	if terraformConfig.Standalone.UpgradeAirgapRancher {
		command = "bash -c '/tmp/non-auth-registry.sh " + terraformConfig.StandaloneRegistry.RegistryName + " " +
			rke2NonAuthRegistryPublicDNS + " " + terraformConfig.Standalone.UpgradedRancherTagVersion + " " +
			terraformConfig.StandaloneRegistry.UpgradedAssetsPath + " " + terraformConfig.Standalone.OSUser + " " +
			terraformConfig.Standalone.UpgradedRancherImage

		if terraformConfig.Standalone.UpgradedRancherAgentImage != "" {
			command += " " + terraformConfig.Standalone.UpgradedRancherAgentImage
		}
	} else {
		command = "bash -c '/tmp/non-auth-registry.sh " + terraformConfig.StandaloneRegistry.RegistryName + " " +
			rke2NonAuthRegistryPublicDNS + " " + terraformConfig.Standalone.RancherTagVersion + " " +
			terraformConfig.StandaloneRegistry.AssetsPath + " " + terraformConfig.Standalone.OSUser + " " +
			terraformConfig.Standalone.RancherImage

		if terraformConfig.Standalone.RancherAgentImage != "" {
			command += " " + terraformConfig.Standalone.RancherAgentImage
		}
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(registryScriptContent) + "' > /tmp/non-auth-registry.sh"),
		cty.StringVal("chmod +x /tmp/non-auth-registry.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// CreateEcrRegistry is a helper function that will create an authenticated ECR registry.
func CreateEcrRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2EcrRegistryPublicDNS string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/", terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/ecr-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2EcrRegistryPublicDNS, ecrRegistry)

	command := "bash -c '/tmp/ecr-registry.sh " + terraformConfig.StandaloneRegistry.ECRURI + " " +
		terraformConfig.Standalone.RancherTagVersion + " " + terraformConfig.Standalone.RancherImage + " " +
		terraformConfig.Standalone.OSUser + " " + terraformConfig.StandaloneRegistry.AssetsPath

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(registryScriptContent) + "' > /tmp/ecr-registry.sh"),
		cty.StringVal("chmod +x /tmp/ecr-registry.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
