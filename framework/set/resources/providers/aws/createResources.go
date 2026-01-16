package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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
		instances = []string{serverOne, serverTwo, serverThree}
		for _, instance := range instances {
			CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
			rootBody.AppendNewline()
		}
	}

	CreateAWSLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	if terraformConfig.Standalone.CertManagerVersion != "" {
		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateTargetGroupAttachments(rootBody, terraformConfig, aws.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, false), port)
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

	instances = []string{serverOne, serverTwo, serverThree}
	for _, instance := range instances {
		CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	CreateAWSLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	if terraformConfig.Standalone.CertManagerVersion != "" {
		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateInternalTargetGroupAttachments(rootBody, terraformConfig, aws.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, true), port)
			rootBody.AppendNewline()
		}

		CreateInternalLoadBalancer(rootBody, terraformConfig)
		rootBody.AppendNewline()

		for _, port := range ports {
			CreateInternalTargetGroups(rootBody, terraformConfig, port)
			rootBody.AppendNewline()

			CreateInternalLoadBalancerListeners(rootBody, port)
			rootBody.AppendNewline()
		}

		CreateRoute53InternalRecord(rootBody, terraformConfig)
		rootBody.AppendNewline()
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// CreateIPv6AWSResources is a helper function that will create the AWS resources needed for the IPv6 RKE2 cluster.
func CreateIPv6AWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateAWSTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	instances = []string{serverOne, serverTwo, serverThree}
	for _, instance := range instances {
		CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	CreateAWSLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	if terraformConfig.Standalone.CertManagerVersion != "" {
		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateTargetGroupAttachments(rootBody, terraformConfig, aws.LoadBalancerTargetGroupAttachment, getTargetGroupAttachment(port, false), port)
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

// getTargetGroupAttachment gets the target group attachment based on the port
func getTargetGroupAttachment(port int64, internal bool) string {
	switch port {
	case 80:
		if internal {
			return aws.InternalTargetGroup80Attachment
		}
		return aws.TargetGroup80Attachment
	case 443:
		if internal {
			return aws.InternalTargetGroup443Attachment
		}
		return aws.TargetGroup443Attachment
	case 6443:
		if internal {
			return aws.InternalTargetGroup6443Attachment
		}
		return aws.TargetGroup6443Attachment
	case 9345:
		if internal {
			return aws.InternalTargetGroup9345Attachment
		}
		return aws.TargetGroup9345Attachment
	default:
		return ""
	}
}
