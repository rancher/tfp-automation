package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	sanity "github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/sirupsen/logrus"
)

const (
	locals            = "locals"
	requiredProviders = "required_providers"
	registry          = "registry"
	rke2Bastion       = "rke2_bastion"
	rke2ServerOne     = "rke2_server1"
	rke2ServerTwo     = "rke2_server2"
	rke2ServerThree   = "rke2_server3"
)

// CreateAWSResources is a helper function that will create the AWS resources needed for the RKE2 cluster.
func CreateAWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig) (*os.File, error) {
	sanity.CreateTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	sanity.CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	instances := []string{rke2Bastion, registry}
	for _, instance := range instances {
		sanity.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	instances = []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}
	for _, instance := range instances {
		CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	sanity.CreateLocalBlock(rootBody)
	rootBody.AppendNewline()

	ports := []int64{80, 443, 6443, 9345}
	for _, port := range ports {
		sanity.CreateTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, sanity.GetTargetGroupAttachment(port), port)
		rootBody.AppendNewline()

		sanity.CreateInternalTargetGroupAttachments(rootBody, defaults.LoadBalancerTargetGroupAttachment, getInternalTargetGroupAttachment(port), port)
		rootBody.AppendNewline()
	}

	sanity.CreateLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	sanity.CreateInternalLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, port := range ports {
		sanity.CreateTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		sanity.CreateInternalTargetGroups(rootBody, terraformConfig, port)
		rootBody.AppendNewline()

		sanity.CreateLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()

		sanity.CreateInternalLoadBalancerListeners(rootBody, port)
		rootBody.AppendNewline()
	}

	sanity.CreateRoute53Record(rootBody, terraformConfig)
	rootBody.AppendNewline()

	sanity.CreateRoute53InternalRecord(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}

// getTargetGroupAttachment gets the target group attachment based on the port
func getInternalTargetGroupAttachment(port int64) string {
	switch port {
	case 80:
		return defaults.InternalTargetGroup80Attachment
	case 443:
		return defaults.InternalTargetGroup443Attachment
	case 6443:
		return defaults.InternalTargetGroup6443Attachment
	case 9345:
		return defaults.InternalTargetGroup9345Attachment
	default:
		return ""
	}
}
