package provisioning

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// setCreateRKE2K3SSnapshot is a function that will set the etcd_snapshot_create resource block in the main.tf file for a RKE2/K3S cluster.
func setCreateRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body) {
	createSnapshotBlock := rkeConfigBlockBody.AppendNewBlock("etcd_snapshot_create", nil)
	createSnapshotBlockBody := createSnapshotBlock.Body()

	generation := int64(1)

	if createSnapshotBlockBody.GetAttribute("generation") == nil {
		createSnapshotBlockBody.SetAttributeValue("generation", cty.NumberIntVal(generation))
	} else {
		createSnapshotBlockBody.SetAttributeValue("generation", cty.NumberIntVal(generation+1))
	}
}

// setRestoreRKE2K3SSnapshot is a function that will set the etcd_snapshot_restore resource block in the main.tf file for a RKE2/K3S cluster.
func setRestoreRKE2K3SSnapshot(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body, snapshots config.Snapshots) {
	restoreSnapshotBlock := rkeConfigBlockBody.AppendNewBlock("etcd_snapshot_restore", nil)
	restoreSnapshotBlockBody := restoreSnapshotBlock.Body()

	generation := int64(1)

	if restoreSnapshotBlockBody.GetAttribute("generation") == nil {
		restoreSnapshotBlockBody.SetAttributeValue("generation", cty.NumberIntVal(generation))
	} else {
		restoreSnapshotBlockBody.SetAttributeValue("generation", cty.NumberIntVal(generation+1))
	}

	restoreSnapshotBlockBody.SetAttributeValue(("name"), cty.StringVal(snapshots.SnapshotName))
	restoreSnapshotBlockBody.SetAttributeValue(("restore_rke_config"), cty.StringVal(snapshots.SnapshotRestore))
}
