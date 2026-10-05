package aws

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setCIEnv(t *testing.T, env map[string]string) {
	for _, key := range []string{"GITHUB_ACTIONS", "GITHUB_REPOSITORY", "GITHUB_WORKFLOW", "GITHUB_WORKFLOW_REF", "GITHUB_RUN_ID", "JOB_NAME", "BUILD_NUMBER"} {
		t.Setenv(key, env[key])
	}
}

func TestCITags(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected map[string]string
	}{
		{
			name: "github actions",
			env: map[string]string{
				"GITHUB_ACTIONS":      "true",
				"GITHUB_REPOSITORY":   "rancher/tests",
				"GITHUB_WORKFLOW":     "Cluster Provisioning",
				"GITHUB_WORKFLOW_REF": "rancher/tests/.github/workflows/cluster-provisioning.yml@refs/heads/main",
				"GITHUB_RUN_ID":       "123456",
			},
			expected: map[string]string{CIJobTag: "rancher/tests/cluster-provisioning.yml", CIBuildTag: "123456"},
		},
		{
			name:     "jenkins",
			env:      map[string]string{"JOB_NAME": "folder/my job (nightly)", "BUILD_NUMBER": "42"},
			expected: map[string]string{CIJobTag: "folder/my job -nightly-", CIBuildTag: "42"},
		},
		{
			name:     "local run",
			env:      map[string]string{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCIEnv(t, tt.env)
			assert.Equal(t, tt.expected, CITags())
		})
	}
}

func TestSetCIDefaultTags(t *testing.T) {
	setCIEnv(t, map[string]string{"JOB_NAME": "folder/job", "BUILD_NUMBER": "7"})

	file := hclwrite.NewEmptyFile()
	providerBody := file.Body().AppendNewBlock("provider", []string{Aws}).Body()
	SetCIDefaultTags(providerBody)

	rendered := string(file.Bytes())
	_, diags := hclsyntax.ParseConfig(file.Bytes(), "main.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors(), diags.Error())
	assert.Contains(t, rendered, "default_tags {")
	assert.Contains(t, rendered, `ci_job   = "folder/job"`)
	assert.Contains(t, rendered, `ci_build = "7"`)
	assert.Equal(t, "ci_build,7,ci_job,folder/job", CIMachineConfigTags())
}

func TestSetCIDefaultTagsWithoutCI(t *testing.T) {
	setCIEnv(t, map[string]string{})

	file := hclwrite.NewEmptyFile()
	providerBody := file.Body().AppendNewBlock("provider", []string{Aws}).Body()
	SetCIDefaultTags(providerBody)

	assert.NotContains(t, string(file.Bytes()), "default_tags")
	assert.Equal(t, "", CIMachineConfigTags())
}
