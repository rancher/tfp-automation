package nestedModules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terratest/modules/terraform"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

// CreateNestedModules is a helper function that creates nested module directories for the given test and module name.
func CreateNestedModules(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options,
	testName, moduleName string) (string, *terraform.Options, error) {
	nestedRancherModuleDirName := namegen.AppendRandomString(strings.ReplaceAll(testName, "_", "-"))
	userDir, _ := rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	parentDir := filepath.Join(userDir, terratestConfig.PathToRepo, moduleName)
	nestedRancherModuleDir := filepath.Join(parentDir, nestedRancherModuleDirName)

	err := os.MkdirAll(nestedRancherModuleDir, os.ModePerm)
	if err != nil {
		return "", nil, err
	}

	perTestTerraformOptions := *terraformOptions
	perTestTerraformOptions.TerraformDir = nestedRancherModuleDir

	return nestedRancherModuleDir, &perTestTerraformOptions, nil
}
