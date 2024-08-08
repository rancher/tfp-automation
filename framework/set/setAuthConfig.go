package set

import (
	"os"

	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ad"
	"github.com/rancher/tfp-automation/framework/set/authproviders/azureAD"
	"github.com/rancher/tfp-automation/framework/set/authproviders/github"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ldap"
	"github.com/rancher/tfp-automation/framework/set/authproviders/okta"
	"github.com/rancher/tfp-automation/framework/set/resources"

	"github.com/sirupsen/logrus"
)

// AuthConfig is a function that will set the main.tf file based on the auth provider.
func AuthConfig(terraformConfig *config.TerraformConfig) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig(configs.Rancher, rancherConfig)

	authProvider := terraformConfig.AuthProvider

	var file *os.File
	keyPath := resources.SetKeyPath()

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	defer file.Close()

	newFile, rootBody := resources.SetProvidersAndUsersTF(rancherConfig, terraformConfig, true)

	rootBody.AppendNewline()

	switch {
	case authProvider == authproviders.AD:
		err = ad.SetAD(terraformConfig, newFile, rootBody, file)
		return err
	case authProvider == authproviders.AzureAD:
		err = azureAD.SetAzureAD(rancherConfig, terraformConfig, newFile, rootBody, file)
		return err
	case authProvider == authproviders.GitHub:
		err = github.SetGithub(terraformConfig, newFile, rootBody, file)
		return err
	case authProvider == authproviders.Okta:
		err = okta.SetOkta(rancherConfig, terraformConfig, newFile, rootBody, file)
		return err
	case authProvider == authproviders.OpenLDAP:
		err = ldap.SetOpenLDAP(terraformConfig, newFile, rootBody, file)
		return err
	default:
		logrus.Errorf("Unsupported auth provider: %v", authProvider)
	}

	return nil
}
