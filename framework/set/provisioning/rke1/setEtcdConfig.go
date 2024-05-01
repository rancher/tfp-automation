package rke1

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setEtcdConfig is a function that will set the etcd configurations in the main.tf file.
func setEtcdConfig(servicesBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	etcdBlock := servicesBlockBody.AppendNewBlock(defaults.Etcd, nil)
	etcdBlockBody := etcdBlock.Body()

	backupConfigBlock := etcdBlockBody.AppendNewBlock(backupConfig, nil)
	backupConfigBlockBody := backupConfigBlock.Body()

	backupConfigBlockBody.SetAttributeValue(defaults.Enabled, cty.BoolVal(true))
	backupConfigBlockBody.SetAttributeValue(intervalHours, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.IntervalHours))
	backupConfigBlockBody.SetAttributeValue(safeTimestamp, cty.BoolVal(terraformConfig.ETCDRKE1.BackupConfig.SafeTimestamp))
	backupConfigBlockBody.SetAttributeValue(timeout, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.Timeout))

	if terraformConfig.Module == modules.EC2RKE1 && terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig != nil {
		s3ConfigBlock := backupConfigBlockBody.AppendNewBlock(s3BackupConfig, nil)
		s3ConfigBlockBody := s3ConfigBlock.Body()

		s3ConfigBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.AccessKey))
		s3ConfigBlockBody.SetAttributeValue(bucketName, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.BucketName))
		s3ConfigBlockBody.SetAttributeValue(defaults.Endpoint, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Endpoint))
		s3ConfigBlockBody.SetAttributeValue(defaults.Folder, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Folder))
		s3ConfigBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Region))
		s3ConfigBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.SecretKey))
	}

	etcdBlockBody.SetAttributeValue(retention, cty.StringVal(terraformConfig.ETCDRKE1.Retention))
	etcdBlockBody.SetAttributeValue(snapshot, cty.BoolVal(false))

	return nil
}
