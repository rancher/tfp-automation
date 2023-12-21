package framework

import (
	"fmt"
	"strings"

	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
)

// SetResourceNodepoolValidation is a function that will validate the nodepool configurations.
func SetResourceNodepoolValidation(pool config.Nodepool, poolNum string) (bool, error) {
	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	module := terraformConfig.Module

	switch {
	case module == aks || module == gke:
		if pool.Quantity <= 0 {
			return false, fmt.Errorf(`Invalid quantity specified for pool %v. Quantity must be greater than 0.`, poolNum)
		}

		return true, nil
	case module == eks:
		if pool.DesiredSize <= 0 {
			return false, fmt.Errorf(`Invalid desired size specified for pool %v. Desired size must be greater than 0.`, poolNum)
		}

		return true, nil
	case strings.Contains(module, rke1) || strings.Contains(module, rke2) || strings.Contains(module, k3s):
		if !pool.Etcd && !pool.Controlplane && !pool.Worker {
			return false, fmt.Errorf(`No roles selected for pool %v. At least one role is required`, poolNum)
		}

		if pool.Quantity <= 0 {
			return false, fmt.Errorf(`Invalid quantity specified for pool %v. Quantity must be greater than 0.`, poolNum)
		}

		return true, nil
	default:
		return false, fmt.Errorf("Unsupported module: %v", module)
	}
}
