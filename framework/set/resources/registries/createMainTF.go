package registries

import (
	"os"
	"sync"
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
	authRegistryPublicDNS    = "auth_registry_public_dns"
	nonAuthRegistryPublicDNS = "non_auth_registry_public_dns"
	globalRegistryPublicDNS  = "global_registry_public_dns"
	ecrRegistryPublicDNS     = "ecr_registry_public_dns"

	authRegistry    = "auth_registry"
	nonAuthRegistry = "non_auth_registry"
	globalRegistry  = "global_registry"
	ecrRegistry     = "ecr_registry"

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

	instances := []string{serverOne, serverTwo, serverThree, authRegistry, nonAuthRegistry, globalRegistry, ecrRegistry}

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

	authRegistryPublicDNS := terraform.Output(t, terraformOptions, authRegistryPublicDNS)
	nonAuthRegistryPublicDNS := terraform.Output(t, terraformOptions, nonAuthRegistryPublicDNS)
	globalRegistryPublicDNS := terraform.Output(t, terraformOptions, globalRegistryPublicDNS)
	ecrRegistryPublicDNS := terraform.Output(t, terraformOptions, ecrRegistryPublicDNS)
	serverOnePublicDNS := terraform.Output(t, terraformOptions, serverOnePublicDNS)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPublicDNS := terraform.Output(t, terraformOptions, serverTwoPublicDNS)
	serverThreePublicDNS := terraform.Output(t, terraformOptions, serverThreePublicDNS)

	// Will create the authenticated registry, unauthenticated registry, and global registry in parallel using goroutines.
	var wg sync.WaitGroup
	var mutex sync.Mutex
	wg.Add(4)

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Creating authenticated registry...")
		file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, authRegistryPublicDNS)
		if err != nil {
			logrus.Fatalf("Error creating authenticated registry: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Creating non-authenticated registry...")
		file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, nonAuthRegistryPublicDNS, nonAuthRegistry)
		if err != nil {
			logrus.Fatalf("Error creating unauthenticated registry: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Creating global registry...")
		file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, globalRegistryPublicDNS, globalRegistry)
		if err != nil {
			logrus.Fatalf("Error creating global registry: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Creating ecr registry...")
		file, err = registry.CreateECRRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, ecrRegistryPublicDNS)
		if err != nil {
			logrus.Fatalf("Error creating ecr registry: %v", err)
		}
	}()

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registries. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", "", err
	}

	wg.Wait()

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

	return authRegistryPublicDNS, nonAuthRegistryPublicDNS, globalRegistryPublicDNS, nil
}
