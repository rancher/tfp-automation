package registries

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	registry "github.com/rancher/tfp-automation/framework/set/resources/registries/createRegistry"
	"github.com/rancher/tfp-automation/framework/set/resources/registries/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/registries/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
)

const (
	authRegistryPublicDNS         = "auth_registry_public_dns"
	unauthRegistryPublicDNS       = "unauth_registry_public_dns"
	authGlobalRegistryPublicDNS   = "auth_global_registry_public_dns"
	unauthGlobalRegistryPublicDNS = "unauth_global_registry_public_dns"
	ecrRegistryPublicDNS          = "ecr_registry_public_dns"

	authRegistryRoute53FQDN         = "auth_registry_route_53_fqdn"
	authGlobalRegistryRoute53FQDN   = "auth_global_registry_route_53_fqdn"
	unauthGlobalRegistryRoute53FQDN = "unauth_global_registry_route_53_fqdn"

	authRegistry         = "auth"
	unauthRegistry       = "unauth"
	authGlobalRegistry   = "auth-global"
	unauthGlobalRegistry = "unauth-global"
	ecrRegistry          = "ecr"

	serverOne            = "server1"
	serverTwo            = "server2"
	serverThree          = "server3"
	serverOnePublicDNS   = "server1_public_dns"
	serverOnePrivateIP   = "server1_private_ip"
	serverTwoPublicDNS   = "server2_public_dns"
	serverThreePublicDNS = "server3_public_dns"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, rancherConfig *shepherdConfig.Config,
	terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, string, string, error) {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{serverOne, serverTwo, serverThree, authRegistry, unauthRegistry, authGlobalRegistry, unauthGlobalRegistry, ecrRegistry}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	var authGlobalRegistryPublicDNS, authGlobalRegistryRoute53FQDN, unauthGlobalRegistryPublicDNS, unauthGlobalRegistryRoute53FQDN string

	if terraformConfig.StandaloneRegistry.CreateAuthGlobalRegistry {
		authGlobalRegistryPublicDNS = terraform.Output(t, terraformOptions, authGlobalRegistryPublicDNS)
		authGlobalRegistryRoute53FQDN = terraform.Output(t, terraformOptions, authGlobalRegistryRoute53FQDN)
	} else if terraformConfig.StandaloneRegistry.CreateUnauthGlobalRegistry {
		unauthGlobalRegistryPublicDNS = terraform.Output(t, terraformOptions, unauthGlobalRegistryPublicDNS)
		unauthGlobalRegistryRoute53FQDN = terraform.Output(t, terraformOptions, unauthGlobalRegistryRoute53FQDN)
	}

	authRegistryPublicDNS := terraform.Output(t, terraformOptions, authRegistryPublicDNS)
	unauthRegistryPublicDNS := terraform.Output(t, terraformOptions, unauthRegistryPublicDNS)
	authRegistryRoute53FQDN := terraform.Output(t, terraformOptions, authRegistryRoute53FQDN)
	ecrRegistryPublicDNS := terraform.Output(t, terraformOptions, ecrRegistryPublicDNS)
	serverOnePublicDNS := terraform.Output(t, terraformOptions, serverOnePublicDNS)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPublicDNS := terraform.Output(t, terraformOptions, serverTwoPublicDNS)
	serverThreePublicDNS := terraform.Output(t, terraformOptions, serverThreePublicDNS)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating unauthenticated registry...")
	file, err = registry.CreateUnauthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, unauthRegistryPublicDNS, unauthRegistry, unauthGlobalRegistryRoute53FQDN, false)
	if err != nil {
		logrus.Fatalf("Error creating unauthenticated registry: %v", err)
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registries. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating global registry...")
	if !terraformConfig.StandaloneRegistry.UseAuthGlobalRegistry {
		file, err = registry.CreateUnauthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, unauthGlobalRegistryPublicDNS, unauthGlobalRegistry, unauthGlobalRegistryRoute53FQDN, true)
		if err != nil {
			logrus.Fatalf("Error creating global registry: %v", err)
		}
	} else {
		file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, authGlobalRegistryPublicDNS, authGlobalRegistry, authGlobalRegistryRoute53FQDN, true)
		if err != nil {
			logrus.Fatalf("Error creating global registry: %v", err)
		}
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registries. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating authenticated registry...")
	file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, authRegistryPublicDNS, authRegistry, authRegistryRoute53FQDN, true)
	if err != nil {
		logrus.Fatalf("Error creating authenticated registry: %v", err)
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registries. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating ecr registry...")
	file, err = registry.CreateECRRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, ecrRegistryPublicDNS)
	if err != nil {
		logrus.Fatalf("Error creating ecr registry: %v", err)
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registries. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	// Needed so that when building the local cluster and setting up Rancher, we use the correct registry images based on
	// whether the global registry is authenticated or not.
	var globalRegistryPublicDNS string
	if !terraformConfig.StandaloneRegistry.UseAuthGlobalRegistry {
		globalRegistryPublicDNS = unauthGlobalRegistryRoute53FQDN
	} else {
		globalRegistryPublicDNS = authGlobalRegistryRoute53FQDN
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, serverOnePublicDNS, serverOnePrivateIP,
		serverTwoPublicDNS, serverThreePublicDNS, globalRegistryPublicDNS)
	if err != nil {
		return "", "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating RKE2 cluster. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, terratestConfig, serverOnePublicDNS, globalRegistryPublicDNS)
	if err != nil {
		return "", "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	return authRegistryRoute53FQDN, unauthRegistryPublicDNS, globalRegistryPublicDNS, nil
}
