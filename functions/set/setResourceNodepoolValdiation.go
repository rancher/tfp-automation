package functions

import (
	"fmt"
	"testing"

	framework "github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/tfp-automation/config"
)

func SetResourceNodepoolValidation(t *testing.T, pool config.Nodepool, poolNum string) (bool, error) {
	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig("terraform", terraformConfig)

	module := terraformConfig.Module

	switch {
	case module == "aks" || module == "gke":

		if pool.Quantity <= 0 {
			t.Logf("invalid quanity specified for pool %v; quantity must be greater than 0.", poolNum)
			return false, fmt.Errorf(`invalid quantity specified for pool` + poolNum + `; quantity must be greater than 0`)
		}

		return true, nil

	case module == "eks":

		if pool.DesiredSize <= 0 {
			t.Logf("invalid desired size specified for pool %v; desired size must be greater than 0.", poolNum)
			return false, fmt.Errorf(`invalid desired size specified for pool` + poolNum + `; desired size must be greater than 0`)
		}

		return true, nil

	case module == "ec2_rke1" || module == "ec2_rke2" || module == "ec2_k3s" || module == "linode_rke1" || module == "linode_rke2" || module == "linode_k3s":

		if !pool.Etcd && !pool.Controlplane && !pool.Worker {
			t.Logf("no roles selected for pool %v; at least one role is required.", poolNum)
			return false, fmt.Errorf(`no roles selected for pool` + poolNum + `; at least one role is required`)
		}

		if pool.Quantity <= 0 {
			t.Logf("invalid quanity specified for pool %v; quantity must be greater than 0.", poolNum)
			return false, fmt.Errorf(`invalid quantity specified for pool` + poolNum + `; quantity must be greater than 0`)
		}

		return true, nil

	default:
		return false, fmt.Errorf("invalid module provided")
	}
}