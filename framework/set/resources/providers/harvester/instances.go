package harvester

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/crypto/ssh"
)

// getPublicSSHKey gets the public key from a private SSH key and returns it as a string
func getPublicSSHKey(privateKeyPath string) string {
	sshBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(sshBytes)
	if err != nil {
		log.Fatalf("Failed to parse OpenSSH private key: %v", err)
		return ""
	}

	publicKey := signer.PublicKey()

	// Convert the public key to authorized_keys format
	pubKeyString := string(ssh.MarshalAuthorizedKey(publicKey))

	return pubKeyString
}

// CreateHarvesterInstances is a function that will set the Harvester instances configurations in the main.tf file.
func CreateHarvesterInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {

	configBlockSSHKey := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.HarvesterSSHKey, hostnamePrefix + "ssh_key"})
	configBlockSSHKeyBody := configBlockSSHKey.Body()

	configBlockSSHKeyBody.SetAttributeValue(defaults.LowerCaseName, cty.StringVal(namegenerator.AppendRandomString("tfpsshkey")))
	configBlockSSHKeyBody.SetAttributeValue(defaults.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))

	publicKey := getPublicSSHKey(terraformConfig.PrivateKeyPath)
	configBlockSSHKeyBody.SetAttributeValue(defaults.PublicKey, cty.StringVal(publicKey))

	configBlockSecret := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.KubernetesSecret, hostnamePrefix + "secret"})
	configBlockSecretBody := configBlockSecret.Body()

	secretBlockMeta := configBlockSecretBody.AppendNewBlock("metadata", []string{})
	secretBlockMetaBody := secretBlockMeta.Body()

	secretName := namegenerator.AppendRandomString("tfpsecret")

	secretBlockMetaBody.SetAttributeValue(defaults.LowerCaseName, cty.StringVal(secretName))
	secretBlockMetaBody.SetAttributeValue(defaults.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))

	secretBlockMetaBody.SetAttributeValue(defaults.Labels, cty.ObjectVal(map[string]cty.Value{
		"sensitive": cty.StringVal("false"),
	}))

	hclLocalValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("{\"userdata\" = local." + defaults.CloudInit + "}")},
	}

	configBlockSecretBody.SetAttributeRaw("data", hclLocalValue)

	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.HarvesterVirtualMachine, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, "custom") {
		configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.NodeCount))
	}

	configBlockBody.AppendNewline()

	formattedList := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`[` + fmt.Sprintf("kubernetes_secret.%s", hostnamePrefix+"secret") + `]`)},
	}

	configBlockBody.SetAttributeRaw(defaults.DependsOn, formattedList)

	randName := namegenerator.AppendRandomString("tfp-vm")
	configBlockBody.SetAttributeValue(defaults.LowerCaseName, cty.StringVal(randName))
	configBlockBody.SetAttributeValue(defaults.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))
	configBlockBody.SetAttributeValue(defaults.RestartAfterUpdate, cty.BoolVal(true))
	configBlockBody.SetAttributeValue(defaults.Description, cty.StringVal(randName))

	tagMap := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("{%s = \"%s\"}", defaults.SshUser, terraformConfig.HarvesterConfig.SSHUser))},
	}

	configBlockBody.SetAttributeRaw(defaults.Tags, tagMap)

	configBlockBody.SetAttributeValue(defaults.CPU, cty.StringVal(terraformConfig.HarvesterConfig.CPUCount))

	configBlockBody.SetAttributeValue(defaults.Memory, cty.StringVal(terraformConfig.HarvesterConfig.MemorySize+defaults.Gi))

	configBlockBody.SetAttributeValue(defaults.EFI, cty.BoolVal(true))
	configBlockBody.SetAttributeValue(defaults.SecureBoot, cty.BoolVal(false))

	configBlockBody.SetAttributeValue(defaults.RunStrategy, cty.StringVal(defaults.RerunOnFailure))
	configBlockBody.SetAttributeValue(defaults.Hostname, cty.StringVal(randName))
	configBlockBody.SetAttributeValue(defaults.MachineType, cty.StringVal(defaults.Q35))

	networkBlock := configBlockBody.AppendNewBlock(defaults.NetworkInterface, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(defaults.LowerCaseName, cty.StringVal(defaults.NIC1))
	networkBlockBody.SetAttributeValue(defaults.WaitForLease, cty.BoolVal(true))
	networkBlockBody.SetAttributeValue(defaults.Model, cty.StringVal(defaults.Virtio))
	networkBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Bridge))
	networkBlockBody.SetAttributeValue(defaults.NetworkName, cty.StringVal(terraformConfig.HarvesterConfig.NetworkNames[0]))

	diskBlock := configBlockBody.AppendNewBlock(defaults.Disk, nil)
	diskBlockBody := diskBlock.Body()

	diskBlockBody.SetAttributeValue(defaults.LowerCaseName, cty.StringVal(defaults.RootDisk))
	diskBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Disk))
	diskBlockBody.SetAttributeValue(defaults.Size, cty.StringVal(terraformConfig.HarvesterConfig.DiskSize+defaults.Gi))
	diskBlockBody.SetAttributeValue(defaults.Bus, cty.StringVal(defaults.Virtio))
	diskBlockBody.SetAttributeValue(defaults.BootOrder, cty.NumberIntVal(1))
	diskBlockBody.SetAttributeValue(defaults.Image, cty.StringVal(terraformConfig.HarvesterConfig.ImageName))
	diskBlockBody.SetAttributeValue(defaults.AutoDelete, cty.BoolVal(true))

	cloudInitBlock := configBlockBody.AppendNewBlock(defaults.CloudInit, nil)
	cloudInitBlockBody := cloudInitBlock.Body()

	cloudInitBlockBody.SetAttributeValue(defaults.UserDataSecretName, cty.StringVal(secretName))
	cloudInitBlockBody.SetAttributeValue(defaults.NetworkData, cty.StringVal(""))

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))

	hostExpression := defaults.Self + "." + "network_interface[0].ip_address"
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal("120"))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}
