package harvester

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// CreateHarvesterResources is a helper function that will create the Harvester resources needed for the RKE2 cluster.
func CreateHarvesterResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateHarvesterProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateHarvesterInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	rootBody.AppendNewline()
	dirList := file.Name()

	localFile, err := os.Create(strings.Join(strings.Split(dirList, "main.tf"), "/") + "/local.yaml")
	if err != nil {
		logrus.Infof("Failed create local.yaml kubeconfig. Error: %v", err)
		return nil, err
	}

	_, err = localFile.Write([]byte(terraformConfig.HarvesterCredentials.KubeconfigContent))
	if err != nil {
		logrus.Infof("Failed write to local.yaml. Error: %v", err)
		return nil, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}
	return file, err
}
