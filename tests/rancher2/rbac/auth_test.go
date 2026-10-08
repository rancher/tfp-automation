//go:build validation

package rbac

import (
	"testing"

	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/rancher/tfp-automation/tests/rancher2/resources"
	"github.com/sirupsen/logrus"
)

func TestTfpAuthConfig(t *testing.T) {
	r := resources.Setup(t)

	tests := []struct {
		name         string
		authProvider string
	}{
		{"Azure_AD", authproviders.AzureAD},
		{"GitHub", authproviders.GitHub},
		{"Okta", authproviders.Okta},
		{"OpenLDAP", authproviders.OpenLDAP},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.TerratestConfig)
		defer file.Close()

		rancher, terraform, _, _ := config.LoadTFPConfigs(r.CattleConfig)
		rancher.AdminToken = r.Client.RancherConfig.AdminToken
		terraform.AuthProvider = tt.authProvider

		terraform = provisioning.UniquifyTerraform(terraform)

		t.Run(tt.name, func(t *testing.T) {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.TerratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(t, r.TerraformOptions, keyPath)

			rbac.AuthConfig(t, rancher, terraform, r.TerraformOptions, []map[string]any{r.CattleConfig}, newFile, rootBody, file)
		})

		params := tfpQase.GetProvisioningSchemaParams(r.TerraformConfig, r.TerratestConfig)
		err := qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.TerratestConfig.LocalQaseReporting {
		results.ReportTest(r.TerratestConfig)
	}
}

func TestTfpAuthConfigDynamicInput(t *testing.T) {
	r := resources.Setup(t)

	if r.TerraformConfig.AuthProvider == "" {
		t.Skip("No auth provider specified")
	}

	tests := []struct {
		name string
	}{
		{r.TerraformConfig.AuthProvider},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.TerratestConfig)
		defer file.Close()

		rancher, terraform, _, _ := config.LoadTFPConfigs(r.CattleConfig)
		rancher.AdminToken = r.Client.RancherConfig.AdminToken
		terraform.AuthProvider = r.TerraformConfig.AuthProvider

		terraform = provisioning.UniquifyTerraform(terraform)

		t.Run(tt.name, func(t *testing.T) {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.TerratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(t, r.TerraformOptions, keyPath)

			rbac.AuthConfig(t, rancher, terraform, r.TerraformOptions, []map[string]any{r.CattleConfig}, newFile, rootBody, file)
		})

		params := tfpQase.GetProvisioningSchemaParams(r.TerraformConfig, r.TerratestConfig)
		err := qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.TerratestConfig.LocalQaseReporting {
		results.ReportTest(r.TerratestConfig)
	}
}
