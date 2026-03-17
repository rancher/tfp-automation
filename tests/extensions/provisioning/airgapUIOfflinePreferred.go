package provisioning

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

const (
	name               = "name"
	uiOfflinePreferred = "ui-offline-preferred"
	value              = "value"
)

// AirgapUIOfflinePreferredCheck is a function that will attempt to load the V3 and VI API pages
// in an airgapped environment.
func AirgapUIOfflinePreferred(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rootBody *hclwrite.Body, newFile *hclwrite.File, file *os.File, setting string, configMap []map[string]any) {
	newFile.Body().Clear()

	if !strings.Contains(string(newFile.Bytes()), general.RequiredProviders) {
		newFile, rootBody = resources.SetProvidersAndUsersTF(rancherConfig, "admin", rancherConfig.AdminPassword, false, newFile, rootBody, configMap, false)
	}

	settingBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.Setting, rancher2.Setting})
	settingBlockBody := settingBlock.Body()

	settingBlockBody.SetAttributeValue(name, cty.StringVal(uiOfflinePreferred))
	settingBlockBody.SetAttributeValue(value, cty.StringVal(setting))

	_, keyPath := resources.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	file, err := os.Create(keyPath + configs.MainTF)
	require.NoError(t, err)

	_, err = file.Write(newFile.Bytes())
	require.NoError(t, err)

	terraform.Init(t, terraformOptions)

	uiOfflinePreferredSettingImport(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	getPageStatus(t, rancherConfig, setting)
}

func uiOfflinePreferredSettingImport(t *testing.T, terraformOptions *terraform.Options) {
	cmd := exec.Command("terraform", "-chdir="+terraformOptions.TerraformDir, "import", rancher2.Setting+"."+rancher2.Setting, uiOfflinePreferred)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Run())
}

func getPageStatus(t *testing.T, rancherConfig *rancher.Config, setting string) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// We want to check both the /v3 and /v1 endpoints. Dynamic and local (true) should be reachable;
	// remote (false) should not be reachable.
	endpoints := []string{"/v3", "/v1"}
	for _, endpoint := range endpoints {
		url := "https://" + rancherConfig.Host + "/api-ui/1.1.11/ui.min.js"
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+rancherConfig.AdminToken)
		start := time.Now()
		resp, err := client.Do(req)
		elapsed := time.Since(start)

		defer func(resp *http.Response) {
			if resp != nil {
				resp.Body.Close()
			}
		}(resp)

		if setting == "dynamic" || setting == "true" {
			require.NoError(t, err, "Error loading ui.min.js for %s", endpoint)
			require.Equal(t, 200, resp.StatusCode)
			require.True(t, elapsed < 10*time.Second)
		} else if setting == "false" {
			if err != nil {
				require.Error(t, err, "Expected error loading ui.min.js for %s, but got none", endpoint)
			} else {
				require.True(t, elapsed >= 10*time.Second)
			}
		}
	}
}
