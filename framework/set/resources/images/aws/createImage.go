package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	imageResource   = "image"
	scriptResource  = "update_image"
	remoteScript    = "/tmp/update-image.sh"
	provisionerFile = "file"
	source          = "source"
	destination     = "destination"
	imageIDOutput   = "image_id"

	// AMIAddress is the Terraform address of the AMI resource.
	AMIAddress = aws.AMIFromInstance + "." + imageResource
)

func raw(expression string) hclwrite.Tokens {
	return hclwrite.Tokens{{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)}}
}

// CreateAMI appends a null_resource that uploads and runs scriptPath on the given aws_instance, followed by an
// aws_ami_from_instance that snapshots it once the script has succeeded.
func CreateAMI(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	instance, scriptPath, imageName string) (*os.File, error) {
	instanceRef := aws.AwsInstance + "." + instance

	scriptBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, scriptResource})
	scriptBody := scriptBlock.Body()

	connectionBody := scriptBody.AppendNewBlock(general.Connection, nil).Body()
	connectionBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	connectionBody.SetAttributeRaw(general.Host, raw(instanceRef+"."+general.PublicIp))
	connectionBody.SetAttributeRaw(general.PrivateKey, raw(general.File+`("`+terraformConfig.PrivateKeyPath+`")`))
	connectionBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
	scriptBody.AppendNewline()

	fileBody := scriptBody.AppendNewBlock(general.Provisioner, []string{provisionerFile}).Body()
	fileBody.SetAttributeValue(source, cty.StringVal(scriptPath))
	fileBody.SetAttributeValue(destination, cty.StringVal(remoteScript))
	scriptBody.AppendNewline()

	execBody := scriptBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec}).Body()
	execBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("chmod +x " + remoteScript),
		cty.StringVal("sudo " + remoteScript),
	}))
	rootBody.AppendNewline()

	amiBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.AMIFromInstance, imageResource})
	amiBody := amiBlock.Body()
	amiBody.SetAttributeValue(general.ResourceName, cty.StringVal(imageName))
	amiBody.SetAttributeRaw(aws.SourceInstanceID, raw(instanceRef+".id"))
	amiBody.SetAttributeRaw(general.DependsOn, raw("["+general.NullResource+"."+scriptResource+"]"))
	rootBody.AppendNewline()

	outputBody := rootBody.AppendNewBlock("output", []string{imageIDOutput}).Body()
	outputBody.SetAttributeRaw(general.Value, raw(AMIAddress+".id"))

	if _, err := file.Write(newFile.Bytes()); err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
