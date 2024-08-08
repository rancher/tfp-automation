package azureAD

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	azureADConfig = "rancher2_auth_config_azuread"

	applicationID     = "application_id"
	applicationSecret = "application_secret"
	authEndpoint      = "auth_endpoint"
	graphEndpoint     = "graph_endpoint"
	rancherURL        = "rancher_url"
	resource          = "resource"
	tenantID          = "tenant_id"
	tokenEndpoint     = "token_endpoint"
)

// SetAzureAD is a function that will set the Azure AD configurations in the main.tf file.
func SetAzureAD(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) error {
	azureADBlock := rootBody.AppendNewBlock(resource, []string{azureADConfig, azureADConfig})
	azureADBlockBody := azureADBlock.Body()

	azureADBlockBody.SetAttributeValue(applicationID, cty.StringVal(terraformConfig.AzureADConfig.ApplicationID))
	azureADBlockBody.SetAttributeValue(applicationSecret, cty.StringVal(terraformConfig.AzureADConfig.ApplicationSecret))
	azureADBlockBody.SetAttributeValue(authEndpoint, cty.StringVal(terraformConfig.AzureADConfig.AuthEndpoint))
	azureADBlockBody.SetAttributeValue(graphEndpoint, cty.StringVal(terraformConfig.AzureADConfig.GraphEndpoint))
	azureADBlockBody.SetAttributeValue(rancherURL, cty.StringVal("https://"+rancherConfig.Host))
	azureADBlockBody.SetAttributeValue(tenantID, cty.StringVal(terraformConfig.AzureADConfig.TenantID))
	azureADBlockBody.SetAttributeValue(tokenEndpoint, cty.StringVal(terraformConfig.AzureADConfig.TokenEndpoint))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Azure AD configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
