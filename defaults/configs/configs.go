package configs

const (
	Rancher   = "rancher"
	Terraform = "terraform"
	Terratest = "terratest"
	TFP       = "tfp"

	DefaultK8sVersion    = "default"
	SecondHighestVersion = "second"

	MainTF          = "/main.tf"
	TerraformFolder = "/.terraform"
	TFState         = "/terraform.tfstate"
	TFStateBackup   = "/terraform.tfstate.backup"
	TFLockHCL       = "/.terraform.lock.hcl"
)
