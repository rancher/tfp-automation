package rke2k3s

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setEtcdConfig is a function that will set the etcd configurations in the main.tf file.
func setEtcdConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	snapshotBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Etcd, nil)
	snapshotBlockBody := snapshotBlock.Body()

	snapshotBlockBody.SetAttributeValue(disableSnapshots, cty.BoolVal(terraformConfig.ETCD.DisableSnapshots))
	snapshotBlockBody.SetAttributeValue(snapshotScheduleCron, cty.StringVal(terraformConfig.ETCD.SnapshotScheduleCron))
	snapshotBlockBody.SetAttributeValue(snapshotRetention, cty.NumberIntVal(int64(terraformConfig.ETCD.SnapshotRetention)))

	if strings.Contains(terraformConfig.Module, modules.EC2) && terraformConfig.ETCD.S3 != nil {
		s3ConfigBlock := snapshotBlockBody.AppendNewBlock(s3Config, nil)
		s3ConfigBlockBody := s3ConfigBlock.Body()

		cloudCredSecretName := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.CloudCredential + "." + defaults.CloudCredential + ".id")},
		}

		s3ConfigBlockBody.SetAttributeValue(bucket, cty.StringVal(terraformConfig.ETCD.S3.Bucket))
		s3ConfigBlockBody.SetAttributeValue(defaults.Endpoint, cty.StringVal(terraformConfig.ETCD.S3.Endpoint))
		s3ConfigBlockBody.SetAttributeRaw(cloudCredentialName, cloudCredSecretName)
		s3ConfigBlockBody.SetAttributeValue(endpointCA, cty.StringVal(terraformConfig.ETCD.S3.EndpointCA))
		s3ConfigBlockBody.SetAttributeValue(defaults.Folder, cty.StringVal(terraformConfig.ETCD.S3.Folder))
		s3ConfigBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.ETCD.S3.Region))
		s3ConfigBlockBody.SetAttributeValue(skipSSLVerify, cty.BoolVal(terraformConfig.ETCD.S3.SkipSSLVerify))
	}

	return nil
}
