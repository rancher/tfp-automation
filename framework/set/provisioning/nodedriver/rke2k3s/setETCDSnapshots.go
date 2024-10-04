package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	EtcdSnapshotCreate  = "etcd_snapshot_create"
	EtcdSnapshotRestore = "etcd_snapshot_restore"

	Generation       = "generation"
	RestoreRKEConfig = "restore_rke_config"
)

// SetCreateRKE2K3SSnapshot is a function that will set the etcd_snapshot_create resource
// block in the main.tf file for a RKE2/K3S cluster.
func SetCreateRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body) {
	createSnapshotBlock := rkeConfigBlockBody.AppendNewBlock(EtcdSnapshotCreate, nil)
	createSnapshotBlockBody := createSnapshotBlock.Body()

	generation := int64(1)

	if createSnapshotBlockBody.GetAttribute(Generation) == nil {
		createSnapshotBlockBody.SetAttributeValue(Generation, cty.NumberIntVal(generation))
	} else {
		createSnapshotBlockBody.SetAttributeValue(Generation, cty.NumberIntVal(generation+1))
	}
}

// SetRestoreRKE2K3SSnapshot is a function that will set the etcd_snapshot_restore
// resource block in the main.tf file for a RKE2/K3S cluster.
func SetRestoreRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body, snapshots config.Snapshots) {
	restoreSnapshotBlock := rkeConfigBlockBody.AppendNewBlock(EtcdSnapshotRestore, nil)
	restoreSnapshotBlockBody := restoreSnapshotBlock.Body()

	generation := int64(1)

	if restoreSnapshotBlockBody.GetAttribute(Generation) == nil {
		restoreSnapshotBlockBody.SetAttributeValue(Generation, cty.NumberIntVal(generation))
	} else {
		restoreSnapshotBlockBody.SetAttributeValue(Generation, cty.NumberIntVal(generation+1))
	}

	restoreSnapshotBlockBody.SetAttributeValue((defaults.ResourceName), cty.StringVal(snapshots.SnapshotName))
	restoreSnapshotBlockBody.SetAttributeValue((RestoreRKEConfig), cty.StringVal(snapshots.SnapshotRestore))
}
