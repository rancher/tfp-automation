package images

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/images"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/standard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	terraformConst = "terraform"
	builder        = "image_builder"
)

// UpdateImage boots a VM from the image configured for the provider (e.g. awsConfig.ami), runs scriptPath on it,
// snapshots the result into a new image and then destroys the VM. The new image is kept. It returns the new image ID.
// If provider is empty, the provider from the terraform config is used. imageName is used as-is, so the caller must pass a unique name.
func UpdateImage(t *testing.T, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, scriptPath, imageName string) (string, error) {
	imageProvider, err := images.TunnelToImageProvider(terraformConfig.Provider)
	require.NoError(t, err)

	scriptPath, err = filepath.Abs(scriptPath)
	require.NoError(t, err)
	_, err = os.Stat(scriptPath)
	require.NoError(t, err, "script not found")

	_, keyPath := rancher2.SetKeyPath(keypath.ImagesKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	var file *os.File
	file = standard.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()
	tfBlockBody := rootBody.AppendNewBlock(terraformConst, nil).Body()

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err = providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, []string{builder})
	require.NoError(t, err)

	// The provider helper already wrote the builder VM to main.tf; rewrite the full file with the image resources added.
	file = standard.OpenFile(file, keyPath)
	logrus.Infof("Running %s on builder VM and creating image %s...", scriptPath, imageName)
	_, err = imageProvider.CreateImage(file, newFile, rootBody, terraformConfig, builder, scriptPath, imageName)
	require.NoError(t, err)

	_, err = terraform.InitAndApplyContextE(t, context.Background(), terraformOptions)
	if err != nil {
		cleanup.Cleanup(t, terraformOptions, keyPath)
		require.NoError(t, err)
	}

	imageID := terraform.OutputContext(t, context.Background(), terraformOptions, images.ImageIDOutput)
	logrus.Infof("Created image %s (%s)", imageName, imageID)

	// Destroying an image resource deregisters the image, so stop tracking it before tearing down the builder VM.
	_, err = terraform.RunTerraformCommandContextE(t, context.Background(), terraformOptions, "state", "rm", imageProvider.ImageResourceAddress)
	require.NoError(t, err)

	cleanup.Cleanup(t, terraformOptions, keyPath)

	return imageID, nil
}
