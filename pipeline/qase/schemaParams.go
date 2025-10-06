package qase

import (
	"strings"
	"time"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	upstream "go.qase.io/qase-api-client"
)

// GetProvisioningSchemaParams gets a set of params from the cattle config and returns a qase params object
func GetProvisioningSchemaParams(configMap map[string]any) []upstream.TestCaseParameterCreate {
	var params []upstream.TestCaseParameterCreate
	var rancherType, upgradedRancherType, amiParam, windows2019AMIParam, windows2022AMIParam upstream.TestCaseParameterCreate

	_, terraform, terratest, _ := config.LoadTFPConfigs(configMap)

	currentDate := time.Now().Format("2006-01-02 03:04PM")

	if terraform.Standalone != nil && terraform.Standalone.RancherImage != "" {
		if terraform.Standalone.RancherImage == "rancher/rancher" {
			rancherType = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Rancher Community", Values: []string{terraform.Standalone.RancherTagVersion}}}
		} else {
			rancherType = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Rancher Prime", Values: []string{terraform.Standalone.RancherTagVersion}}}
		}
	}

	if terraform.Standalone != nil && terraform.Standalone.UpgradedRancherImage != "" {
		if terraform.Standalone.UpgradedRancherImage == "rancher/rancher" {
			upgradedRancherType = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Upgraded to Rancher Community", Values: []string{terraform.Standalone.UpgradedRancherTagVersion}}}
		} else {
			upgradedRancherType = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Upgraded to Rancher Prime", Values: []string{terraform.Standalone.UpgradedRancherTagVersion}}}
		}
	}

	if strings.Contains(terraform.Module, modules.EC2) {
		amiParam = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "AMI", Values: []string{terraform.AWSConfig.AMI}}}
	}

	if strings.Contains(terraform.Module, clustertypes.WINDOWS) && strings.Contains(terraform.Module, "2019") {
		windows2019AMIParam = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Windows2019AMI", Values: []string{terraform.AWSConfig.Windows2019AMI}}}
	}

	if strings.Contains(terraform.Module, clustertypes.WINDOWS) && strings.Contains(terraform.Module, "2022") {
		windows2022AMIParam = upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Windows2022AMI", Values: []string{terraform.AWSConfig.Windows2022AMI}}}
	}

	moduleParam := upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Module", Values: []string{terraform.Module}}}
	k8sParam := upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "K8sVersion", Values: []string{terratest.KubernetesVersion}}}
	cniParam := upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "CNI", Values: []string{terraform.CNI}}}
	timeParam := upstream.TestCaseParameterCreate{ParameterSingle: &upstream.ParameterSingle{Title: "Time", Values: []string{currentDate}}}

	params = append(params, rancherType, upgradedRancherType, amiParam, windows2019AMIParam, windows2022AMIParam, moduleParam, k8sParam, cniParam, timeParam)

	return params
}
