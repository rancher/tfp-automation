package aws

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	DefaultTags = "default_tags"
	Tags        = "tags"
	CIJobTag    = "ci_job"
	CIBuildTag  = "ci_build"

	maxTagValueLength = 256
)

var invalidTagChars = regexp.MustCompile(`[^\p{L}\p{N} _.:/=+\-@]`)

func sanitizeTagValue(value string) string {
	value = invalidTagChars.ReplaceAllString(strings.TrimSpace(value), "-")
	if runes := []rune(value); len(runes) > maxTagValueLength {
		value = string(runes[:maxTagValueLength])
	}

	return value
}

// CITags will return the ci_job and ci_build tags for the Jenkins or GitHub Actions build running this code, or an empty map outside of CI.
func CITags() map[string]string {
	var job, build string

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		workflow := os.Getenv("GITHUB_WORKFLOW")
		if ref := os.Getenv("GITHUB_WORKFLOW_REF"); ref != "" {
			workflow = path.Base(strings.SplitN(ref, "@", 2)[0])
		}

		job = workflow
		if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" && workflow != "" {
			job = repo + "/" + workflow
		}

		build = os.Getenv("GITHUB_RUN_ID")
	} else {
		job = os.Getenv("JOB_NAME")
		build = os.Getenv("BUILD_NUMBER")
	}

	tags := map[string]string{}
	if job = sanitizeTagValue(job); job != "" {
		tags[CIJobTag] = job
	}

	if build = sanitizeTagValue(build); build != "" {
		tags[CIBuildTag] = build
	}

	return tags
}

// SetDefaultTags will add a default_tags block with the given tags to the aws provider block.
func SetDefaultTags(awsProvBlockBody *hclwrite.Body, tags map[string]string) {
	tagValues := map[string]cty.Value{}
	for key, value := range tags {
		tagValues[key] = cty.StringVal(value)
	}

	defaultTagsBlock := awsProvBlockBody.AppendNewBlock(DefaultTags, nil)
	defaultTagsBlock.Body().SetAttributeValue(Tags, cty.MapVal(tagValues))
}

// CIMachineConfigTags will return the CI tags in the "key1,value1,key2,value2" format used by amazonec2_config, or an empty string outside of CI.
func CIMachineConfigTags() string {
	tags := CITags()

	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	pairs := make([]string, 0, len(tags)*2)
	for _, key := range keys {
		pairs = append(pairs, key, tags[key])
	}

	return strings.Join(pairs, ",")
}
