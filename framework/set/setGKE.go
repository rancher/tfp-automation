package framework

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetGKE is a function that will set the GKE configurations in the main.tf file.
func SetGKE(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig("terraform", terraformConfig)

	googleAuthEncodedJSONConfig := new(config.GoogleAuthEncodedJSON)
	framework.LoadConfig("googleAuthEncodedJSON", googleAuthEncodedJSONConfig)

	newFile, rootBody := setProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	googleCredConfigBlock := cloudCredBlockBody.AppendNewBlock("google_credential_config", nil)

	authEncodedJSON := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`jsonencode({"type" = "` + googleAuthEncodedJSONConfig.Type + `", "project_id" = "` + googleAuthEncodedJSONConfig.ProjectID + `", "private_key_id" = "` + googleAuthEncodedJSONConfig.PrivateKeyID + `", "private_key" = "` + googleAuthEncodedJSONConfig.PrivateKey + `", "client_email" = "` + googleAuthEncodedJSONConfig.ClientEmail + `", "client_id" = "` + googleAuthEncodedJSONConfig.ClientID + `", "auth_uri" = "` + googleAuthEncodedJSONConfig.AuthURI + `", "token_uri" = "` + googleAuthEncodedJSONConfig.TokenURI + `", "auth_provider_x509_cert_url" = "` + googleAuthEncodedJSONConfig.AuthProviderX509CertURL + `", "client_x509_cert_url" = "` + googleAuthEncodedJSONConfig.ClientX509CertURL + `"})`)},
	}

	googleCredConfigBlock.Body().SetAttributeRaw("auth_encoded_json", authEncodedJSON)

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster", "rancher2_cluster"})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))

	gkeConfigBlock := clusterBlockBody.AppendNewBlock("gke_config_v2", nil)
	gkeConfigBlockBody := gkeConfigBlock.Body()

	gkeConfigBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))

	cloudCredSecret := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_cloud_credential.rancher2_cloud_credential.id`)},
	}

	gkeConfigBlockBody.SetAttributeRaw("google_credential_secret", cloudCredSecret)
	gkeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.Region))
	gkeConfigBlockBody.SetAttributeValue("project_id", cty.StringVal(terraformConfig.GKEProjectID))
	gkeConfigBlockBody.SetAttributeValue("kubernetes_version", cty.StringVal(k8sVersion))
	gkeConfigBlockBody.SetAttributeValue("network", cty.StringVal(terraformConfig.GKENetwork))
	gkeConfigBlockBody.SetAttributeValue("subnetwork", cty.StringVal(terraformConfig.GKESubnetwork))

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		SetResourceNodepoolValidation(pool, poolNum)

		nodePoolsBlock := gkeConfigBlockBody.AppendNewBlock("node_pools", nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue("initial_node_count", cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue("max_pods_constraint", cty.NumberIntVal(pool.MaxPodsContraint))
		nodePoolsBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.HostnamePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue("version", cty.StringVal(k8sVersion))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write GKE configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
