package qase

import (
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	upstream "go.qase.io/qase-api-client"
)

// GetProvisioningSchemaParams gets a set of params from the cattle config and returns a qase params object
func GetProvisioningSchemaParams(configMap map[string]any) []upstream.TestCaseParameterCreate {
	var params []upstream.TestCaseParameterCreate
	_, terraform, terratest, _ := config.LoadTFPConfigs(configMap)

	params = append(params,
		getRunType(terraform),
		getRancherType(terraform),
		getCNIParam(terraform),
		getAMIParam(terraform),
		getWindowsAMIParam(terraform),
		getK8sParam(terratest),
		getTurtlesParam(terraform),
	)

	return params
}

func getRunType(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	var version string
	if terraform.Standalone != nil && terraform.Standalone.UpgradedRancherTagVersion == "" {
		version = terraform.Standalone.RancherTagVersion
	} else if terraform.Standalone != nil && terraform.Standalone.UpgradedRancherTagVersion != "" {
		version = terraform.Standalone.UpgradedRancherTagVersion
	}

	switch {
	case terraform.Standalone == nil:
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Scheduled Testing", Values: []string{""}}}
	case version != "" && !strings.Contains(version, "-alpha") && !strings.Contains(version, "-rc"):
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Scheduled Testing", Values: []string{version}}}
	case version != "" && strings.Contains(version, "-alpha"):
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Release Testing", Values: []string{version}}}
	case version != "" && strings.Contains(version, "-rc"):
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "RC Testing", Values: []string{version}}}
	}

	return upstream.TestCaseParameterCreate{}
}

func getRancherType(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	var image, version, title string
	isUpgrade := false
	prevImage := ""
	prevVersion := ""

	if terraform.Standalone != nil && terraform.Standalone.UpgradedRancherImage == "" && terraform.Standalone.UpgradedRancherTagVersion == "" {
		image = terraform.Standalone.RancherImage
		version = terraform.Standalone.RancherTagVersion
	} else if terraform.Standalone != nil && terraform.Standalone.UpgradedRancherImage != "" && terraform.Standalone.UpgradedRancherTagVersion != "" {
		image = terraform.Standalone.UpgradedRancherImage
		version = terraform.Standalone.UpgradedRancherTagVersion
		prevImage = terraform.Standalone.RancherImage
		prevVersion = terraform.Standalone.RancherTagVersion
		isUpgrade = true
	}

	if image != "" {
		if isUpgrade {
			var fromType, toType string
			switch prevImage {
			case "rancher/rancher":
				fromType = "Rancher Community"
			default:
				fromType = "Rancher Prime"
			}

			switch image {
			case "rancher/rancher":
				toType = "Rancher Community"
			default:
				toType = "Rancher Prime"
			}

			title = "Upgraded From " + fromType + ": " + prevVersion + " to " + toType
		} else {
			switch image {
			case "rancher/rancher":
				title = "Rancher Community"
			default:
				title = "Rancher Prime"
			}
		}

		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: title, Values: []string{version}}}
	}

	return upstream.TestCaseParameterCreate{}
}

func getCNIParam(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "CNI", Values: []string{terraform.CNI}}}
}

func getAMIParam(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	if strings.Contains(terraform.Module, modules.EC2) {
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "AMI", Values: []string{terraform.AWSConfig.AMI}}}
	}

	return upstream.TestCaseParameterCreate{}
}

func getWindowsAMIParam(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	if strings.Contains(terraform.Module, clustertypes.WINDOWS) && strings.Contains(terraform.Module, "2019") {
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Windows2019AMI", Values: []string{terraform.AWSConfig.Windows2019AMI}}}
	}

	if strings.Contains(terraform.Module, clustertypes.WINDOWS) && strings.Contains(terraform.Module, "2022") {
		return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Windows2022AMI", Values: []string{terraform.AWSConfig.Windows2022AMI}}}
	}

	return upstream.TestCaseParameterCreate{}
}

func getK8sParam(terratest *config.TerratestConfig) upstream.TestCaseParameterCreate {
	return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "K8sVersion", Values: []string{terratest.KubernetesVersion}}}
}

func getTurtlesParam(terraform *config.TerraformConfig) upstream.TestCaseParameterCreate {
	var prevTurtles, turtles, upgradedTurtles, title, value string

	if terraform.Standalone != nil && terraform.Standalone.FeatureFlags.UpgradedTurtles == "" {
		turtles = terraform.Standalone.FeatureFlags.Turtles
	} else if terraform.Standalone != nil && terraform.Standalone.FeatureFlags.UpgradedTurtles != "" {
		upgradedTurtles = terraform.Standalone.FeatureFlags.UpgradedTurtles
		prevTurtles = terraform.Standalone.FeatureFlags.Turtles
	}

	if terraform.Standalone.FeatureFlags.UpgradedTurtles != "" {
		title = "Turtles status pre-upgrade: " + prevTurtles + " to Turtles status post-upgrade: "
		value = upgradedTurtles
	} else {
		title = "Turtles status: "
		value = turtles
	}

	return upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: title, Values: []string{value}}}
}
