package functions

import (
	"testing"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	framework "github.com/rancher/rancher/tests/framework/pkg/config"
	waitAction "github.com/rancher/tfp-automation/functions/wait/action"
	waitState "github.com/rancher/tfp-automation/functions/wait/state"
	"github.com/rancher/tfp-automation/config"
)

func WaitFor(t *testing.T, client *rancher.Client, clusterID string, action string) {
	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig("terraform", terraformConfig)

	module := terraformConfig.Module

	if module == "aks" || module == "eks" || module == "ec2_k3s" || module == "ec2_rke1" || module == "ec2_rke2" || module == "linode_k3s" || module == "linode_rke1" || module == "linode_rke2" {
		if module != "eks" && !((module == "ec2_rke1" || module == "linode_rke1") && action == "kubernetes-upgrade") {
			waitState.WaitingOrUpdating(t, client, clusterID)
		}

		waitState.ActiveAndReady(t, client, clusterID)

		if action == "scale-up" || action == "kubernetes-upgrade" {
			waitState.ActiveNodes(t, client, clusterID)

			waitState.ActiveAndReady(t, client, clusterID)
		}
	}

	if action == "scale-up" {
		waitAction.ScaleUp(t, client, clusterID)
		waitState.ActiveAndReady(t, client, clusterID)
	}

	if action == "scale-down" {
		waitAction.ScaleDown(t, client, clusterID)
		waitState.ActiveAndReady(t, client, clusterID)
	}

	if action == "kubernetes-upgrade" {
		waitAction.KubernetesUpgrade(t, client, clusterID, module)
		waitState.ActiveAndReady(t, client, clusterID)
	}

}
