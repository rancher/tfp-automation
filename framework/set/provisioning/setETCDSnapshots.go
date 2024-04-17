package provisioning

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/snapshot"
	"github.com/zclconf/go-cty/cty"
)

// setCreateRKE2K3SSnapshot is a function that will set the etcd_snapshot_create resource
// block in the main.tf file for a RKE2/K3S cluster.
func setCreateRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body) {
	createSnapshotBlock := rkeConfigBlockBody.AppendNewBlock(snapshot.EtcdSnapshotCreate, nil)
	createSnapshotBlockBody := createSnapshotBlock.Body()

	generation := int64(1)

	if createSnapshotBlockBody.GetAttribute(snapshot.Generation) == nil {
		createSnapshotBlockBody.SetAttributeValue(snapshot.Generation, cty.NumberIntVal(generation))
	} else {
		createSnapshotBlockBody.SetAttributeValue(snapshot.Generation, cty.NumberIntVal(generation+1))
	}
}

// setRestoreRKE2K3SSnapshot is a function that will set the etcd_snapshot_restore
// resource block in the main.tf file for a RKE2/K3S cluster.
func setRestoreRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body, snapshots config.Snapshots) {
	restoreSnapshotBlock := rkeConfigBlockBody.AppendNewBlock(snapshot.EtcdSnapshotRestore, nil)
	restoreSnapshotBlockBody := restoreSnapshotBlock.Body()

	generation := int64(1)

	if restoreSnapshotBlockBody.GetAttribute(snapshot.Generation) == nil {
		restoreSnapshotBlockBody.SetAttributeValue(snapshot.Generation, cty.NumberIntVal(generation))
	} else {
		restoreSnapshotBlockBody.SetAttributeValue(snapshot.Generation, cty.NumberIntVal(generation+1))
	}

	restoreSnapshotBlockBody.SetAttributeValue((resourceblocks.ResourceName), cty.StringVal(snapshots.SnapshotName))
	restoreSnapshotBlockBody.SetAttributeValue((snapshot.RestoreRKEConfig), cty.StringVal(snapshots.SnapshotRestore))
}
