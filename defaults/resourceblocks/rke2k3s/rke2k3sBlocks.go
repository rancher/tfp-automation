package rke2k3s

const (
	ClusterV2       = "rancher2_cluster_v2"
	MachineConfigV2 = "rancher2_machine_config_v2"

	CloudCredentialName       = "cloud_credential_name"
	CloudCredentialSecretName = "cloud_credential_secret_name"
	ControlPlaneRole          = "control_plane_role"
	EtcdRole                  = "etcd_role"
	WorkerRole                = "worker_role"

	UpgradeStrategy         = "upgrade_strategy"
	ControlPlaneConcurrency = "control_plane_concurrency"
	WorkerConcurrency       = "worker_concurrency"

	DisableSnapshots     = "disable_snapshots"
	SnapshotScheduleCron = "snapshot_schedule_cron"
	SnapshotRetention    = "snapshot_retention"
	S3Config             = "s3_config"
	Bucket               = "bucket"
	EndpointCA           = "endpoint_ca"
	SkipSSLVerify        = "skip_ssl_verify"
)
