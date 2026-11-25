package images

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/images/aws"
)

// ImageResourceFunc appends the resources that run a script on the already created builder instance and then
// snapshot that instance into a new image. It must also emit an output named ImageIDOutput holding the new image's ID.
type ImageResourceFunc func(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	instance, scriptPath, imageName string) (*os.File, error)

type ImageResources struct {
	CreateImage ImageResourceFunc
	// ImageResourceAddress is the Terraform address of the image resource. It is removed from state before destroy so
	// that tearing down the builder VM does not delete the image that was just created.
	ImageResourceAddress string
}

const ImageIDOutput = "image_id"

// TunnelToImageProvider returns the helpers needed to create an image on a given provider.
// To support a new provider, add a package that implements ImageResourceFunc and register it here.
func TunnelToImageProvider(provider string) (ImageResources, error) {
	switch provider {
	case providers.AWS:
		return ImageResources{CreateImage: aws.CreateAMI, ImageResourceAddress: aws.AMIAddress}, nil
	default:
		return ImageResources{}, fmt.Errorf("image updates are not yet supported for provider: %s", provider)
	}
}
