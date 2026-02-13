package cluster

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
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

	switch {
	case terraformConfig.Provider == providers.AKS:
		aksScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/hosted/cluster/create-aks-cluster.sh")

		aksScriptContent, err := os.ReadFile(aksScriptPath)
		if err != nil {
			return nil, err
		}

		aksPublicKey, err := os.ReadFile(terraformConfig.AzureConfig.KeyPath)
		if err != nil {
			return nil, err
		}

		encodedPUBFile := base64.StdEncoding.EncodeToString([]byte(aksPublicKey))

		createAKSCluster(rootBody, terraformConfig, bastionPublicIP, aksScriptContent, encodedPUBFile)
	case terraformConfig.Provider == providers.EKS:
		eksScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/hosted/cluster/create-eks-cluster.sh")

		eksScriptContent, err := os.ReadFile(eksScriptPath)
		if err != nil {
			return nil, err
		}

		createEKSCluster(rootBody, terraformConfig, bastionPublicIP, eksScriptContent)
	case terraformConfig.Provider == providers.GKE:
		gkeScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/hosted/cluster/create-gke-cluster.sh")

		gkeScriptContent, err := os.ReadFile(gkeScriptPath)
		if err != nil {
			return nil, err
		}

		encodedJson := base64.StdEncoding.EncodeToString([]byte(terraformConfig.GoogleCredentials.AuthEncodedJSON))
		createGKECluster(rootBody, terraformConfig, bastionPublicIP, gkeScriptContent, encodedJson)
	}

	_, err := file.Write(newFile.Bytes())
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

// createEKSCluster is a helper function that will create the local EKS cluster server.
func createEKSCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, bastionPublicIP string, script []byte) {
	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, bastionPublicIP, hostedCluster)

	command := "bash -c '/tmp/create-eks-cluster.sh " + terraformConfig.ResourcePrefix + " " + terraformConfig.AWSConfig.Region + " " +
		terraformConfig.AWSCredentials.AWSAccessKey + " " + terraformConfig.AWSCredentials.AWSSecretKey + "'"

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/create-eks-cluster.sh"),
		cty.StringVal("chmod +x /tmp/create-eks-cluster.sh"),
		cty.StringVal(command),
	}))
}

// createGKECluster is a helper function that will create the local GKE cluster server.
func createGKECluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, bastionPublicIP string, script []byte,
	encodedJson string) {
	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, bastionPublicIP, hostedCluster)

	command := "bash -c '/tmp/create-gke-cluster.sh " + terraformConfig.ResourcePrefix + " " + terraformConfig.GoogleConfig.Zone + " " +
		terraformConfig.GoogleConfig.MachineType + " " + terraformConfig.GoogleConfig.ProjectID + " " +
		encodedJson + "'"

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/create-gke-cluster.sh"),
		cty.StringVal("chmod +x /tmp/create-gke-cluster.sh"),
		cty.StringVal(command),
	}))
}
