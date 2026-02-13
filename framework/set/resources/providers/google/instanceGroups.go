package google

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

// CreateGoogleCloudInstanceGroups will set up the Google Cloud instance groups.
func CreateGoogleCloudInstanceGroups(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	instanceGroupBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeInstanceGroup, googleDefaults.GoogleComputeInstanceGroup + "_" + strconv.FormatInt(port, 10)})
	instanceGroupBlockBody := instanceGroupBlock.Body()

	instanceGroupBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-group-"+strconv.FormatInt(port, 10)))
	instanceGroupBlockBody.SetAttributeValue(googleDefaults.GoogleZone, cty.StringVal(terraformConfig.GoogleConfig.Zone))

	expression := fmt.Sprintf("[" + googleDefaults.GoogleComputeInstance + `.` + server1 + `.self_link, ` +
		googleDefaults.GoogleComputeInstance + `.` + server2 + `.self_link, ` +
		googleDefaults.GoogleComputeInstance + `.` + server3 + `.self_link]`)

	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	instanceGroupBlockBody.SetAttributeRaw(instances, value)

	namedPortBlock := instanceGroupBlockBody.AppendNewBlock(namedPort, nil)
	namedPortBlockBody := namedPortBlock.Body()

	namedPortBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-port-"+strconv.FormatInt(port, 10)))
	namedPortBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
}
