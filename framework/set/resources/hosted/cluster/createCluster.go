package cluster

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
	hostedCluster = "hostedCluster"
)

// CreateHostedCluster is a helper function that will create the local hosted cluster.
func CreateHostedCluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	bastionPublicIP string, terratestConfig *config.TerratestConfig) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RKE2KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	aksScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/hosted/cluster/create-aks-cluster.sh")

	aksScriptContent, err := os.ReadFile(aksScriptPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := os.ReadFile(terraformConfig.AzureConfig.KeyPath)
	if err != nil {
		return nil, err
	}

	encodedPUBFile := base64.StdEncoding.EncodeToString([]byte(publicKey))

	switch {
	case terraformConfig.LocalHostedCluster.AKS:
		createAKSCluster(rootBody, terraformConfig, bastionPublicIP, aksScriptContent, encodedPUBFile)
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// createAKSCluster is a helper function that will create the local AKS cluster server.
func createAKSCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, bastionPublicIP string, script []byte,
	encodedPUBFile string) {
	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, bastionPublicIP, hostedCluster)

	command := "bash -c '/tmp/create-aks-cluster.sh " + terraformConfig.ResourcePrefix + " " + terraformConfig.AzureConfig.Location + " " +
		terraformConfig.AzureConfig.AKSNodeCount + " " + terraformConfig.AzureConfig.VMSize + " " + terraformConfig.AzureCredentials.ClientID + " " +
		terraformConfig.AzureCredentials.ClientSecret + " " + terraformConfig.AzureCredentials.TenantID + " " + terraformConfig.AzureConfig.SSHUser + " " +
		encodedPUBFile + "'"

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/create-aks-cluster.sh"),
		cty.StringVal("chmod +x /tmp/create-aks-cluster.sh"),
		cty.StringVal(command),
	}))
}
