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
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/harvester"
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

	configBlockSSHKey := rootBody.AppendNewBlock(general.Resource, []string{harvester.HarvesterSSHKey, hostnamePrefix + "ssh_key"})
	configBlockSSHKeyBody := configBlockSSHKey.Body()

	configBlockSSHKeyBody.SetAttributeValue(harvester.LowerCaseName, cty.StringVal(namegenerator.AppendRandomString("tfpsshkey")))
	configBlockSSHKeyBody.SetAttributeValue(general.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))

	publicKey := getPublicSSHKey(terraformConfig.PrivateKeyPath)
	configBlockSSHKeyBody.SetAttributeValue(harvester.PublicKey, cty.StringVal(publicKey))
	configBlockSecret := rootBody.AppendNewBlock(general.Resource, []string{harvester.KubernetesSecret, hostnamePrefix + "secret"})
	configBlockSecretBody := configBlockSecret.Body()

	secretBlockMeta := configBlockSecretBody.AppendNewBlock("metadata", []string{})
	secretBlockMetaBody := secretBlockMeta.Body()

	secretName := namegenerator.AppendRandomString("tfpsecret")

	secretBlockMetaBody.SetAttributeValue(harvester.LowerCaseName, cty.StringVal(secretName))
	secretBlockMetaBody.SetAttributeValue(general.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))

	secretBlockMetaBody.SetAttributeValue(harvester.Labels, cty.ObjectVal(map[string]cty.Value{
		"sensitive": cty.StringVal("false"),
	}))

	hclLocalValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("{\"userdata\" = local." + harvester.CloudInit + "}")},
	}

	configBlockSecretBody.SetAttributeRaw("data", hclLocalValue)

	configBlock := rootBody.AppendNewBlock(general.Resource, []string{harvester.HarvesterVirtualMachine, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		configBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(totalNodeCount))
	}

	configBlockBody.AppendNewline()

	formattedList := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`[` + fmt.Sprintf("kubernetes_secret.%s", hostnamePrefix+"secret") + `]`)},
	}

	configBlockBody.SetAttributeRaw(general.DependsOn, formattedList)

	randName := namegenerator.AppendRandomString("tfp-vm")
	configBlockBody.SetAttributeValue(harvester.LowerCaseName, cty.StringVal(randName))
	configBlockBody.SetAttributeValue(general.Namespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))
	configBlockBody.SetAttributeValue(harvester.RestartAfterUpdate, cty.BoolVal(true))
	configBlockBody.SetAttributeValue(general.Description, cty.StringVal(randName))

	tagMap := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("{%s = \"%s\"}", harvester.SshUser, terraformConfig.HarvesterConfig.SSHUser))},
	}

	configBlockBody.SetAttributeRaw(general.Tags, tagMap)

	configBlockBody.SetAttributeValue(harvester.CPU, cty.StringVal(terraformConfig.HarvesterConfig.CPUCount))

	configBlockBody.SetAttributeValue(harvester.Memory, cty.StringVal(terraformConfig.HarvesterConfig.MemorySize+harvester.Gi))

	configBlockBody.SetAttributeValue(harvester.EFI, cty.BoolVal(true))
	configBlockBody.SetAttributeValue(harvester.SecureBoot, cty.BoolVal(false))

	configBlockBody.SetAttributeValue(harvester.RunStrategy, cty.StringVal(harvester.RerunOnFailure))
	configBlockBody.SetAttributeValue(harvester.Hostname, cty.StringVal(randName))
	configBlockBody.SetAttributeValue(harvester.MachineType, cty.StringVal(harvester.Q35))
	networkBlock := configBlockBody.AppendNewBlock(harvester.NetworkInterface, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(harvester.LowerCaseName, cty.StringVal(harvester.NIC1))
	networkBlockBody.SetAttributeValue(harvester.WaitForLease, cty.BoolVal(true))
	networkBlockBody.SetAttributeValue(harvester.Model, cty.StringVal(harvester.Virtio))
	networkBlockBody.SetAttributeValue(general.Type, cty.StringVal(harvester.Bridge))
	networkBlockBody.SetAttributeValue(harvester.NetworkName, cty.StringVal(terraformConfig.HarvesterConfig.NetworkNames[0]))

	diskBlock := configBlockBody.AppendNewBlock(harvester.Disk, nil)
	diskBlockBody := diskBlock.Body()

	diskBlockBody.SetAttributeValue(harvester.LowerCaseName, cty.StringVal(harvester.RootDisk))
	diskBlockBody.SetAttributeValue(general.Type, cty.StringVal(harvester.Disk))
	diskBlockBody.SetAttributeValue(harvester.Size, cty.StringVal(terraformConfig.HarvesterConfig.DiskSize+harvester.Gi))
	diskBlockBody.SetAttributeValue(harvester.Bus, cty.StringVal(harvester.Virtio))
	diskBlockBody.SetAttributeValue(harvester.BootOrder, cty.NumberIntVal(1))
	diskBlockBody.SetAttributeValue(harvester.Image, cty.StringVal(terraformConfig.HarvesterConfig.ImageName))
	diskBlockBody.SetAttributeValue(harvester.AutoDelete, cty.BoolVal(true))

	cloudInitBlock := configBlockBody.AppendNewBlock(harvester.CloudInit, nil)
	cloudInitBlockBody := cloudInitBlock.Body()

	cloudInitBlockBody.SetAttributeValue(harvester.UserDataSecretName, cty.StringVal(secretName))
	cloudInitBlockBody.SetAttributeValue(harvester.NetworkData, cty.StringVal(""))
	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))

	hostExpression := general.Self + "." + "network_interface[0].ip_address"
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(harvester.Timeout, cty.StringVal("120"))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}
