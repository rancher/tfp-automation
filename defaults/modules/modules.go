package modules

const (
	HostedAzureAKS  = "azure_aks_hosted"
	HostedAWSEKS    = "aws_eks_hosted"
	HostedGoogleGKE = "google_gke_hosted"

	NodeDriverAzureRKE2 = "azure_rke2_nodedriver"
	NodeDriverAzureK3S  = "azure_k3s_nodedriver"

	CustomAWSRKE2            = "aws_rke2_custom"
	CustomAWSRKE2Windows2019 = "aws_rke2_windows_2019_custom"
	CustomAWSRKE2Windows2022 = "aws_rke2_windows_2022_custom"
	CustomAWSK3S             = "aws_k3s_custom"

	CustomVsphereRKE2 = "vsphere_rke2_custom"
	CustomVsphereK3S  = "vsphere_k3s_custom"

	AWS               = "aws"
	NodeDriverAWSRKE2 = "aws_rke2_nodedriver"
	NodeDriverAWSK3S  = "aws_k3s_nodedriver"

	NodeDriverGoogleRKE2 = "google_rke2_nodedriver"
	NodeDriverGoogleK3S  = "google_k3s_nodedriver"

	NodeDriverHarvesterRKE2 = "harvester_rke2_nodedriver"
	NodeDriverHarvesterK3S  = "harvester_k3s_nodedriver"

	ImportedAWSRKE2            = "aws_rke2_imported"
	ImportedAWSRKE2Windows2019 = "aws_rke2_windows_2019_imported"
	ImportedAWSRKE2Windows2022 = "aws_rke2_windows_2022_imported"
	ImportedAWSK3S             = "aws_k3s_imported"

	ImportedVsphereRKE2 = "vsphere_rke2_imported"
	ImportedVsphereK3S  = "vsphere_k3s_imported"

	NodeDriverLinodeRKE2 = "linode_rke2_nodedriver"
	NodeDriverLinodeK3S  = "linode_k3s_nodedriver"

	NodeDriverVsphereRKE2 = "vsphere_rke2_nodedriver"
	NodeDriverVsphereK3S  = "vsphere_k3s_nodedriver"

	AirgapAWSRKE2            = "aws_rke2_airgap"
	AirgapAWSRKE2Windows2019 = "aws_rke2_windows_2019_airgap"
	AirgapAWSRKE2Windows2022 = "aws_rke2_windows_2022_airgap"
	AirgapAWSK3S             = "aws_k3s_airgap"
)
