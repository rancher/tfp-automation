package set

import (
	"os"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ad"
	"github.com/rancher/tfp-automation/framework/set/authproviders/azureAD"
	"github.com/rancher/tfp-automation/framework/set/authproviders/github"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ldap"
	"github.com/rancher/tfp-automation/framework/set/authproviders/okta"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"

	"github.com/sirupsen/logrus"
)

// AuthConfig is a function that will set the main.tf file based on the auth provider.
func AuthConfig(testUser, testPassword string, configMap []map[string]any) error {
	rancherConfig, terraform, _ := config.LoadTFPConfigs(configMap[0])

	authProvider := terraform.AuthProvider

	var file *os.File
	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	defer file.Close()

	newFile, rootBody := resources.SetProvidersAndUsersTF(testUser, testPassword, true, configMap)

	rootBody.AppendNewline()

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
