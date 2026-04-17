package provisioning

import (
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
)

func UniquifyTerraform(terraformConfig *config.TerraformConfig) *config.TerraformConfig {
	terraformConfig.ResourcePrefix = namegen.AppendRandomString(terraformConfig.ResourcePrefix)

	return terraformConfig
}
