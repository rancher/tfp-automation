package rancher2

import (
	"fmt"
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
)

// SetResourceNodepoolValidation is a function that will validate the nodepool configurations.
func SetResourceNodepoolValidation(terraformConfig *config.TerraformConfig, pool config.Nodepool, poolNum string) (bool, error) {
	module := terraformConfig.Module

	switch {
	case module == clustertypes.AKS || module == clustertypes.GKE:
		if pool.Quantity <= 0 {
			return false, fmt.Errorf(`Invalid quantity specified for pool %v. Quantity must be greater than 0.`, poolNum)
		}

		return true, nil
	case module == clustertypes.EKS:
		if pool.DesiredSize <= 0 {
			return false, fmt.Errorf(`Invalid desired size specified for pool %v. Desired size must be greater than 0.`, poolNum)
		}

		return true, nil
	case strings.Contains(module, clustertypes.RKE1) || strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S):
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
