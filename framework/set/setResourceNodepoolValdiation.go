package framework

import (
	"strings"

	ranchFrame "github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// SetResourceNodepoolValidation is a function that will validate the nodepool configurations.
func SetResourceNodepoolValidation(pool config.Nodepool, poolNum string) {
	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	module := terraformConfig.Module

	switch {
	case module == aks || module == gke:
		if pool.Quantity <= 0 {
			logrus.Errorf(`Invalid quantity specified for pool %v. Quantity must be greater than 0.`, poolNum)
		}
	case module == eks:
		if pool.DesiredSize <= 0 {
			logrus.Errorf(`Invalid desired size specified for pool %v. Desired size must be greater than 0.`, poolNum)
		}
	case strings.Contains(module, rke1) || strings.Contains(module, rke2) || strings.Contains(module, k3s):
		if !pool.Etcd && !pool.Controlplane && !pool.Worker {
			logrus.Errorf(`No roles selected for pool %v. At least one role is required`, poolNum)
		}
		if pool.Quantity <= 0 {
			logrus.Errorf(`Invalid quantity specified for pool %v. Quantity must be greater than 0.`, poolNum)
		}
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}
}
