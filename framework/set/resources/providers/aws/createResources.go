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
	CreateAWSTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	if terraformConfig.Proxy != nil {
		instances = []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}
		for _, instance := range instances {
			CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
			rootBody.AppendNewline()
		}
	}

	CreateAWSLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	if terraformConfig.Standalone.RancherHostname != "" {
		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, false), port)
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
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// CreateAirgappedAWSResources is a helper function that will create the AWS resources needed for the airagpped RKE2 cluster.
func CreateAirgappedAWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateAWSTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	instances = []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}
	for _, instance := range instances {
		CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	CreateAWSLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	ports := []int64{80, 443, 6443, 9345}
	for _, port := range ports {
		CreateTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, false), port)
		rootBody.AppendNewline()

		CreateInternalTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, true), port)
		rootBody.AppendNewline()
	}

	CreateLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateInternalLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, port := range ports {
		CreateTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		CreateInternalTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		CreateLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()

		CreateInternalLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()
	}

	CreateRoute53Record(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateRoute53InternalRecord(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// getTargetGroupAttachment gets the target group attachment based on the port
func getTargetGroupAttachment(port int64, internal bool) string {
	switch port {
	case 80:
		if internal {
			return defaults.InternalTargetGroup80Attachment
		}
		return defaults.TargetGroup80Attachment
	case 443:
		if internal {
			return defaults.InternalTargetGroup443Attachment
		}
		return defaults.TargetGroup443Attachment
	case 6443:
		if internal {
			return defaults.InternalTargetGroup6443Attachment
		}
		return defaults.TargetGroup6443Attachment
	case 9345:
		if internal {
			return defaults.InternalTargetGroup9345Attachment
		}
		return defaults.TargetGroup9345Attachment
	default:
		return ""
	}
}
