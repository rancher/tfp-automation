package rke

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// CheckClusterStatus is a helper function that will check the status of the RKE1 cluster.
func CheckClusterStatus(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rkeServerOnePublicIP, kubeConfig string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/rke/rke/cluster.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	encodedKubeConfig := base64.StdEncoding.EncodeToString([]byte(kubeConfig))
	command := fmt.Sprintf("bash -c \"/tmp/cluster.sh '%s'\"", encodedKubeConfig)

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rkeServerOnePublicIP, rkeServerOne)

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(scriptContent) + "' > /tmp/cluster.sh"),
		cty.StringVal("chmod +x /tmp/cluster.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
