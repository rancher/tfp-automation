package rke

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	address          = "address"
	cluster          = "cluster"
	enableCriDockerD = "enable_cri_dockerd"
	rkeServerOne     = "rke_server1"
	rkeServerTwo     = "rke_server2"
	rkeServerThree   = "rke_server3"
	role             = "role"
	sshKey           = "ssh_key"
	user             = "user"
)

// CreateRKECluster is a helper function that will create the RKE2 cluster.
func CreateRKECluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) (*os.File, error) {
	rkeBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.RKECluster, cluster})
	rkeBlockBody := rkeBlock.Body()

	instances := []string{rkeServerOne, rkeServerTwo, rkeServerThree}
	for _, instance := range instances {
		nodesBlock := rkeBlockBody.AppendNewBlock(defaults.Nodes, nil)
		nodesBlockBody := nodesBlock.Body()

		addressExpression := defaults.AwsInstance + "." + instance + ".public_ip"
		values := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(addressExpression)},
		}

		nodesBlockBody.SetAttributeRaw(address, values)
		nodesBlockBody.SetAttributeValue(user, cty.StringVal(terraformConfig.Standalone.RKE1User))

		rolesExpression := fmt.Sprintf(`["controlplane", "etcd", "worker"]`)
		values = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(rolesExpression)},
		}

		nodesBlockBody.SetAttributeRaw(role, values)

		keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		nodesBlockBody.SetAttributeRaw(sshKey, keyPath)
	}

	rkeBlockBody.SetAttributeValue(enableCriDockerD, cty.BoolVal(true))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
