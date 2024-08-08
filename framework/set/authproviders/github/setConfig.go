package github

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	githubConfig = "rancher2_auth_config_github"

	resource     = "resource"
	clientID     = "client_id"
	clientSecret = "client_secret"
)

// SetGithub is a function that will set the Github configurations in the main.tf file.
func SetGithub(terraformConfig *config.TerraformConfig, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	githubBlock := rootBody.AppendNewBlock(resource, []string{githubConfig, githubConfig})
	githubBlockBody := githubBlock.Body()

	githubBlockBody.SetAttributeValue(clientID, cty.StringVal(terraformConfig.GithubConfig.ClientID))
	githubBlockBody.SetAttributeValue(clientSecret, cty.StringVal(terraformConfig.GithubConfig.ClientSecret))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Github configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
