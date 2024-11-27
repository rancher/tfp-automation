package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
)

// CreateAWSResources is a helper function that will create the AWS resources needed for the RKE2 cluster.
func CreateAWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) (*os.File, error) {
	createTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	createAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	instances := []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}
	for _, instance := range instances {
		createAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	createLocalBlock(rootBody)
	rootBody.AppendNewline()

	ports := []int64{80, 443, 6443, 9345}
	for _, port := range ports {
		createTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port), port)
		rootBody.AppendNewline()
	}

	createLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, port := range ports {
		createTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		createLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()
	}

	createRoute53Record(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// getTargetGroupAttachment gets the target group attachment based on the port
func getTargetGroupAttachment(port int64) string {
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
