package set

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ad"
	"github.com/rancher/tfp-automation/framework/set/authproviders/azureAD"
	"github.com/rancher/tfp-automation/framework/set/authproviders/github"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ldap"
	"github.com/rancher/tfp-automation/framework/set/authproviders/okta"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"

	"github.com/sirupsen/logrus"
)

// AuthConfig is a function that will set the main.tf file based on the auth provider.
func AuthConfig(rancherConfig *rancher.Config, testUser, testPassword string, configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	var err error

	newFile, rootBody = resources.SetProvidersAndUsersTF(rancherConfig, testUser, testPassword, true, newFile, rootBody, configMap, false)

	rancherConfig, terraform, _, _ := config.LoadTFPConfigs(configMap[0])
	authProvider := terraform.AuthProvider

	switch {
	case authProvider == authproviders.AD:
		err = ad.SetAD(terraform, newFile, rootBody, file)
		return err
	case authProvider == authproviders.AzureAD:
		err = azureAD.SetAzureAD(rancherConfig, terraform, newFile, rootBody, file)
		return err
	case authProvider == authproviders.GitHub:
		err = github.SetGithub(terraform, newFile, rootBody, file)
		return err
	case authProvider == authproviders.Okta:
		err = okta.SetOkta(rancherConfig, terraform, newFile, rootBody, file)
		return err
	case authProvider == authproviders.OpenLDAP:
		err = ldap.SetOpenLDAP(terraform, newFile, rootBody, file)
		return err
	default:
		logrus.Errorf("Unsupported auth provider: %v", authProvider)
	}

	return nil
}
