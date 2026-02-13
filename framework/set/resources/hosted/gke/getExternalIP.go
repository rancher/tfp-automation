package gke

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rancher/tfp-automation/config"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s %v\nOutput:\n%s\nError: %w", name, args, string(output), err)
	}

	return nil
}

func runCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s %v\nOutput:\n%s\nError: %w", name, args, string(output), err)
	}

	return strings.TrimSpace(string(output)), nil
}

func getIngressExternalIP() (string, error) {
	cmd := exec.Command("sh", "-c", "export PATH=\"$PWD/google-cloud-sdk/bin:$PATH\" && kubectl get service ingress-nginx-controller --namespace=ingress-nginx -o wide | awk 'NR==2 {print $4}'")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("command failed: %s\nOutput:\n%s\nError: %w", cmd.String(), string(output), err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GrabGKEExternalIP is a helper function that will grab the external IP of the LoadBalancer created for the GKE cluster.
func GrabGKEExternalIP(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, error) {
	if err := runCommand("gcloud", "components", "install", "gke-gcloud-auth-plugin", "--quiet"); err != nil {
		return "", err
	}

	if err := runCommand("gcloud", "config", "set", "project", terraformConfig.GoogleConfig.ProjectID); err != nil {
		return "", err
	}

	if err := runCommand("gcloud", "container", "clusters", "get-credentials", terraformConfig.ResourcePrefix, "--region",
		terraformConfig.GoogleConfig.Zone); err != nil {
		return "", err
	}

	externalIP, err := getIngressExternalIP()
	if err != nil {
		return "", err
	}

	return externalIP, nil
}
