package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
)

// CreateAWSResources is a helper function that will create the AWS resources needed for the RKE2 cluster.
func CreateAWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	CreateLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	ports := []int64{80, 443, 6443, 9345}
	for _, port := range ports {
		CreateTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, GetTargetGroupAttachment(port), port)
		rootBody.AppendNewline()
	}

	CreateLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, port := range ports {
		CreateTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		CreateLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()
	}

	CreateRoute53Record(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// GetTargetGroupAttachment gets the target group attachment based on the port
func GetTargetGroupAttachment(port int64) string {
	switch port {
	case 80:
		return defaults.TargetGroup80Attachment
	case 443:
		return defaults.TargetGroup443Attachment
	case 6443:
		return defaults.TargetGroup6443Attachment
	case 9345:
		return defaults.TargetGroup9345Attachment
	default:
		return ""
	}
}
