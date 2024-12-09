package configs

import namegen "github.com/rancher/shepherd/pkg/namegenerator"

const (
	Rancher   = "rancher"
	Terraform = "terraform"
	Terratest = "terratest"
	TFP       = "tfp"

	TestUser     = "testuser"
	TestPassword = "testpassword"

	DefaultK8sVersion    = "default"
	SecondHighestVersion = "second"

	MainTF          = "/main.tf"
	RKEDebugLog     = "/rke_debug.log"
	TerraformFolder = "/.terraform"
	TFState         = "/terraform.tfstate"
	TFStateBackup   = "/terraform.tfstate.backup"
	TFLockHCL       = "/.terraform.lock.hcl"
)

// CreateTestCredentials creates test credentials for the test user, password, cluster name, and pool name.
func CreateTestCredentials() (string, string, string, string) {
	testUser := namegen.AppendRandomString(TestUser)
	testPassword := namegen.AppendRandomString(TestPassword)
	clusterName := namegen.AppendRandomString(TFP)
	poolName := namegen.AppendRandomString(TFP)

	return testUser, testPassword, clusterName, poolName
}
