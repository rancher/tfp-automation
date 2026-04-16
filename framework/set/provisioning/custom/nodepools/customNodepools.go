package nodepools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	rancher2resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

type AWSInstanceGroup struct {
	ResourceName string
	AMI          string
	InstanceType string
	Quantity     int64
}

// TotalNodeCount returns the number of non-Windows nodes across all configured nodepools.
func TotalNodeCount(terratestConfig *config.TerratestConfig) int64 {
	var totalNodeCount int64

	for _, pool := range terratestConfig.Nodepools {
		if pool.Windows {
			continue
		}

		totalNodeCount += pool.Quantity
	}

	return totalNodeCount
}

// WindowsNodeCount returns the total number of Windows nodes across all configured nodepools.
func WindowsNodeCount(terratestConfig *config.TerratestConfig) int64 {
	var totalNodeCount int64

	for _, pool := range terratestConfig.Nodepools {
		if !pool.Windows {
			continue
		}

		totalNodeCount += pool.Quantity
	}

	return totalNodeCount
}

// BuildRoleFlags expands nodepool role flags into one entry per non-Windows node.
func BuildRoleFlags(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) ([]string, error) {
	roleFlags := make([]string, 0, TotalNodeCount(terratestConfig))

	for count, pool := range terratestConfig.Nodepools {
		if pool.Windows {
			continue
		}

		_, err := rancher2resources.SetResourceNodepoolValidation(terraformConfig, pool, strconv.Itoa(count))
		if err != nil {
			return nil, err
		}

		poolRoleFlags := buildPoolRoleFlags(pool)
		for i := int64(0); i < pool.Quantity; i++ {
			roleFlags = append(roleFlags, poolRoleFlags)
		}
	}

	return roleFlags, nil
}

// BuildAWSInstanceGroups converts non-Windows nodepools into AWS instance group definitions.
func BuildAWSInstanceGroups(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) ([]AWSInstanceGroup, error) {
	instanceGroups := make([]AWSInstanceGroup, 0, len(terratestConfig.Nodepools))

	for count, pool := range terratestConfig.Nodepools {
		if pool.Windows {
			continue
		}

		_, err := rancher2resources.SetResourceNodepoolValidation(terraformConfig, pool, strconv.Itoa(count))
		if err != nil {
			return nil, err
		}

		ami := terraformConfig.AWSConfig.AMI
		instanceType := terraformConfig.AWSConfig.AWSInstanceType
		if pool.Worker && terraformConfig.MixedArchitecture {
			ami = terraformConfig.AWSConfig.ARMAMI
			instanceType = terraformConfig.AWSConfig.ARMInstanceType
		}

		instanceGroups = append(instanceGroups, AWSInstanceGroup{
			ResourceName: fmt.Sprintf("%s-pool-%d", terraformConfig.ResourcePrefix, count),
			AMI:          ami,
			InstanceType: instanceType,
			Quantity:     pool.Quantity,
		})
	}

	return instanceGroups, nil
}

// BuildAWSPublicIPExpression builds the Terraform expression for all AWS instance public IPs.
func BuildAWSPublicIPExpression(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, error) {
	instanceGroups, err := BuildAWSInstanceGroups(terraformConfig, terratestConfig)
	if err != nil {
		return "", err
	}

	groupExpressions := make([]string, 0, len(instanceGroups))
	for _, group := range instanceGroups {
		groupExpressions = append(groupExpressions, fmt.Sprintf("%s.%s.*.public_ip", aws.AwsInstance, group.ResourceName))
	}

	return fmt.Sprintf("flatten([%s])", strings.Join(groupExpressions, ", ")), nil
}

func buildPoolRoleFlags(pool config.Nodepool) string {
	roleFlags := make([]string, 0, 3)

	if pool.Etcd {
		roleFlags = append(roleFlags, clusters.EtcdRoleFlag)
	}

	if pool.Controlplane {
		roleFlags = append(roleFlags, clusters.ControlPlaneRoleFlag)
	}

	if pool.Worker {
		roleFlags = append(roleFlags, clusters.WorkerRoleFlag)
	}

	return strings.Join(roleFlags, " ")
}
