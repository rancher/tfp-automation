package createRegistry

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	authRegistry   = "auth_registry"
	globalRegistry = "global_registry"
	unauthRegistry = "unauth_registry"
	ecrRegistry    = "ecr_registry"
)

// CreateAuthenticatedRegistry is a helper function that will create an authenticated registry.
func CreateAuthenticatedRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2AuthRegistryPublicDNS, registryType, rke2AuthRegistryRoute53FQDN string,
	useSecureFQDN bool) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RegistryKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/auth-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	privateFullChain, err := os.ReadFile(terraformConfig.PrivateFullChainPath)
	if err != nil {
		return nil, err
	}

	privateCertKey, err := os.ReadFile(terraformConfig.PrivateCertKeyPath)
	if err != nil {
		return nil, err
	}

	encodedFullChain := base64.StdEncoding.EncodeToString((privateFullChain))
	encodedCertKey := base64.StdEncoding.EncodeToString((privateCertKey))

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2AuthRegistryPublicDNS, registryType)

	command := "/tmp/auth-registry.sh " + terraformConfig.Standalone.CertManagerVersion + " " + terraformConfig.StandaloneRegistry.RegistryName + " " + terraformConfig.StandaloneRegistry.RegistryUsername + " " +
		terraformConfig.StandaloneRegistry.RegistryPassword + " " + terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
		rke2AuthRegistryPublicDNS + " " + terraformConfig.Standalone.RancherTagVersion + " " + terraformConfig.StandaloneRegistry.AssetsPath + " " +
		terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.RancherImage + " " + encodedFullChain + " " + encodedCertKey

	if useSecureFQDN {
		command += " " + rke2AuthRegistryRoute53FQDN
	} else {
		command += " \"\""
	}

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	} else {
		command += " \"\""
	}

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
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

// CreateUnauthenticatedRegistry is a helper function that will create an unauthenticated registry.
func CreateUnauthenticatedRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2UnauthRegistryPublicDNS, registryType, rke2UnauthRegistryRoute53FQDN string,
	useSecureFQDN bool) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RegistryKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/unauth-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	privateFullChain, err := os.ReadFile(terraformConfig.PrivateFullChainPath)
	if err != nil {
		return nil, err
	}

	privateCertKey, err := os.ReadFile(terraformConfig.PrivateCertKeyPath)
	if err != nil {
		return nil, err
	}

	encodedFullChain := base64.StdEncoding.EncodeToString((privateFullChain))
	encodedCertKey := base64.StdEncoding.EncodeToString((privateCertKey))

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2UnauthRegistryPublicDNS, registryType)

	var command string

	if terraformConfig.Standalone.UpgradeAirgapRancher {
		command = "/tmp/unauth-registry.sh " + terraformConfig.StandaloneRegistry.RegistryName + " " + terraformConfig.Standalone.CertManagerVersion + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + rke2UnauthRegistryPublicDNS + " " +
			terraformConfig.Standalone.UpgradedRancherTagVersion + " " + terraformConfig.StandaloneRegistry.UpgradedAssetsPath + " " +
			terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.UpgradedRancherImage + " " + encodedFullChain + " " + encodedCertKey

		if useSecureFQDN {
			command += " " + rke2UnauthRegistryRoute53FQDN
		} else {
			command += " \"\""
		}

		if terraformConfig.Standalone.RancherAgentImage != "" {
			command += " " + terraformConfig.Standalone.UpgradedRancherAgentImage
		} else {
			command += " \"\""
		}
	} else {
		command = "/tmp/unauth-registry.sh " + terraformConfig.StandaloneRegistry.RegistryName + " " + terraformConfig.Standalone.CertManagerVersion + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + rke2UnauthRegistryPublicDNS + " " +
			terraformConfig.Standalone.RancherTagVersion + " " + terraformConfig.StandaloneRegistry.AssetsPath + " " + terraformConfig.Standalone.OSUser + " " +
			terraformConfig.Standalone.RancherImage + " " + encodedFullChain + " " + encodedCertKey

		if useSecureFQDN {
			command += " " + rke2UnauthRegistryRoute53FQDN
		} else {
			command += " \"\""
		}

		if terraformConfig.Standalone.RancherAgentImage != "" {
			command += " " + terraformConfig.Standalone.RancherAgentImage
		} else {
			command += " \"\""
		}
	}

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(registryScriptContent) + "' > /tmp/unauth-registry.sh"),
		cty.StringVal("chmod +x /tmp/unauth-registry.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// CreateECRRegistry is a helper function that will create an authenticated ECR registry.
func CreateECRRegistry(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2EcrRegistryPublicDNS string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RegistryKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/registries/createRegistry/ecr-registry.sh")

	registryScriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2EcrRegistryPublicDNS, ecrRegistry)

	command := "/tmp/ecr-registry.sh " + terraformConfig.StandaloneRegistry.ECRURI + " " + terraformConfig.Standalone.RegistryUsername + " " +
		terraformConfig.Standalone.RegistryPassword + " " + terraformConfig.Standalone.RancherTagVersion + " " + terraformConfig.Standalone.RancherImage + " " +
		terraformConfig.Standalone.OSUser + " " + terraformConfig.StandaloneRegistry.AssetsPath + " " + terraformConfig.AWSCredentials.AWSAccessKey + " " +
		terraformConfig.AWSCredentials.AWSSecretKey + " " + terraformConfig.AWSConfig.Region

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	} else {
		command += " \"\""
	}

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
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
