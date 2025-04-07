package registries

import (
	"os"
	"sync"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
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

	authRegistry    = "auth_registry"
	nonAuthRegistry = "non_auth_registry"
	globalRegistry  = "global_registry"

	rke2ServerOne            = "rke2_server1"
	rke2ServerTwo            = "rke2_server2"
	rke2ServerThree          = "rke2_server3"
	rke2ServerOnePublicDNS   = "rke2_server1_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPublicDNS   = "rke2_server2_public_dns"
	rke2ServerThreePublicDNS = "rke2_server3_public_dns"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratest *config.TerratestConfig) (string, string, string, error) {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree, authRegistry, nonAuthRegistry, globalRegistry}

	providerTunnel := providers.TunnelToProvider(terraformConfig)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratest, instances)
	if err != nil {
		return "", "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	authRegistryPublicDNS := terraform.Output(t, terraformOptions, authRegistryPublicDNS)
	nonAuthRegistryPublicDNS := terraform.Output(t, terraformOptions, nonAuthRegistryPublicDNS)
	globalRegistryPublicDNS := terraform.Output(t, terraformOptions, globalRegistryPublicDNS)
	rke2ServerOnePublicDNS := terraform.Output(t, terraformOptions, rke2ServerOnePublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicDNS := terraform.Output(t, terraformOptions, rke2ServerTwoPublicDNS)
	rke2ServerThreePublicDNS := terraform.Output(t, terraformOptions, rke2ServerThreePublicDNS)

	// Will create the authenticated registry, unauthenticated registry, and global registry in parallel using goroutines.
	var wg sync.WaitGroup
	var mutex sync.Mutex
	wg.Add(3)

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Creating authenticated registry...")
		file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, authRegistryPublicDNS)
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
		file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, nonAuthRegistryPublicDNS, nonAuthRegistry)
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
		file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, globalRegistryPublicDNS, globalRegistry)
		if err != nil {
			logrus.Fatalf("Error creating global registry: %v", err)
		}
	}()

	terraform.InitAndApply(t, terraformOptions)

	wg.Wait()

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicDNS, rke2ServerOnePrivateIP,
		rke2ServerTwoPublicDNS, rke2ServerThreePublicDNS, globalRegistryPublicDNS)
	if err != nil {
		return "", "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicDNS, globalRegistryPublicDNS)
	if err != nil {
		return "", "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	return authRegistryPublicDNS, nonAuthRegistryPublicDNS, globalRegistryPublicDNS, nil
}
