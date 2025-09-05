package qase

import (
	"strings"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	upstream "go.qase.io/client"
)

// GetProvisioningSchemaParams gets a set of params from the cattle config and returns a qase params object
func GetProvisioningSchemaParams(client *rancher.Client, configMap map[string]any) []upstream.Params {
	var params []upstream.Params
	var amiParam upstream.Params

	_, terraform, terratest, _ := config.LoadTFPConfigs(configMap)

	currentDate := time.Now().Format("2006-01-02 03:04PM")

	if strings.Contains(terraform.Module, modules.EC2) {
		amiParam = upstream.Params{Title: "AMI", Values: []string{terraform.AWSConfig.AMI}}
	}

	moduleParam := upstream.Params{Title: "Module", Values: []string{terraform.Module}}
	k8sParam := upstream.Params{Title: "K8sVersion", Values: []string{terratest.KubernetesVersion}}
	cniParam := upstream.Params{Title: "CNI", Values: []string{terraform.CNI}}
	timeParam := upstream.Params{Title: "Time", Values: []string{currentDate}}

	params = append(params, amiParam, moduleParam, k8sParam, cniParam, timeParam)

	return params
}
