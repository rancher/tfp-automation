package rke2k3s

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	customnodepools "github.com/rancher/tfp-automation/framework/set/provisioning/custom/nodepools"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/sleep"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported"
	resources "github.com/rancher/tfp-automation/framework/set/resources/imported"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	rancher2resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

const (
	windows = "windows"
)

// // SetImportedRKE2K3s is a function that will set the imported RKE2/K3s cluster configurations in the main.tf file.
func SetImportedRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	imported.SetImportedCluster(rootBody, terraformConfig.ResourcePrefix)
	rootBody.AppendNewline()

	serverNodeNames, agentNodeNames, err := buildImportedLinuxNodeNames(terraformConfig, terratestConfig)
	if err != nil {
		return nil, nil, err
	}

	linuxNodeNames := append(append([]string{}, serverNodeNames...), agentNodeNames...)
	if len(serverNodeNames) == 0 {
		return nil, nil, fmt.Errorf("at least one non-windows imported nodepool entry is required")
	}

	serverOneName := serverNodeNames[0]
	windowsNodeName := terraformConfig.ResourcePrefix + `-` + windows

	nodePublicIPs, nodePublicIPv6s, nodePrivateIPs := getProviderIPAddresses(terraformConfig, terratestConfig, rootBody, linuxNodeNames)

	token := namegen.AppendRandomString(general.Import)

	if terraformConfig.AWSConfig.ClusterCIDR != "" && !terraformConfig.AWSConfig.IPv6AddressOnly {
		err = resources.CreateDualStackRKE2K3SImportedCluster(rootBody, terraformConfig, terratestConfig, linuxNodeNames, serverNodeNames, agentNodeNames, nodePublicIPs, nodePrivateIPs, token)
	} else if terraformConfig.AWSConfig.ClusterCIDR != "" && terraformConfig.AWSConfig.IPv6AddressOnly {
		err = resources.CreateIPv6RKE2K3SImportedCluster(rootBody, terraformConfig, terratestConfig, linuxNodeNames, serverNodeNames, agentNodeNames, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs, token)
	} else if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		err = resources.CreateProxyRKE2K3SImportedCluster(rootBody, terraformConfig, terratestConfig, linuxNodeNames, serverNodeNames, agentNodeNames, nodePublicIPs, nodePrivateIPs, token)
	} else if terraformConfig.AWSConfig.ClusterCIDR == "" {
		err = resources.CreateRKE2K3SImportedCluster(rootBody, terraformConfig, terratestConfig, linuxNodeNames, serverNodeNames, agentNodeNames, nodePublicIPs, nodePrivateIPs, token)
	}
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
		rootBody.AppendNewline()

		windowsNodePublicDNS := fmt.Sprintf("${%s.%s.public_dns}", awsDefaults.AwsInstance, windowsNodeName)
		resources.AddWindowsNodeToImportedCluster(rootBody, terraformConfig, terratestConfig, nodePrivateIPs[serverOneName], windowsNodePublicDNS, token)

		// Add the sleep command to wait for the Windows node to be ready
		rootBody.AppendNewline()
		dependsOnValue := fmt.Sprintf("[" + general.NullResource + ".add_windows_node" + "]")

		sleep.SetTimeSleep(rootBody, terraformConfig, "10s", dependsOnValue, "import_wins")
		rootBody.AppendNewline()
	}

	importCommand := imported.GetImportCommand(terraformConfig.ResourcePrefix)

	err = imported.ImportNodes(rootBody, terraformConfig, terratestConfig, nodePublicIPs[serverOneName], importCommand[serverOneName], serverNodeNames[1:], agentNodeNames)
	if err != nil {
		return nil, nil, err
	}

	return newFile, file, nil
}

func buildImportedLinuxNodeNames(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) ([]string, []string, error) {
	linuxNodeCount := customnodepools.TotalNodeCount(terratestConfig)
	serverNodes := make([]string, 0, linuxNodeCount)
	agentNodes := make([]string, 0, linuxNodeCount)

	nodeIndex := int64(1)
	for count, pool := range terratestConfig.Nodepools {
		if pool.Windows {
			continue
		}

		if _, err := rancher2resources.SetResourceNodepoolValidation(terraformConfig, pool, strconv.Itoa(count)); err != nil {
			return nil, nil, err
		}

		for i := int64(0); i < pool.Quantity; i++ {
			nodeName := fmt.Sprintf("%s_server%d", terraformConfig.ResourcePrefix, nodeIndex)
			if pool.Etcd || pool.Controlplane {
				serverNodes = append(serverNodes, nodeName)
			} else {
				agentNodes = append(agentNodes, nodeName)
			}

			nodeIndex++
		}
	}

	return serverNodes, agentNodes, nil
}
