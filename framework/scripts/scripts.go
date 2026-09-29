package scripts

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/sirupsen/logrus"
)

const (
	errorPrefix      = "Error: "
	warningPrefix    = "Warning: "
	remoteExecPrefix = " (remote-exec): "
	processExited    = "Process exited with status "
	maxErrorLines    = 5
)

// InitAndApplyE runs terraform init and apply and returns a summary of each Terraform error.
func InitAndApplyE(t *testing.T, terraformOptions *terraform.Options) (string, error) {
	output, err := terraform.InitAndApplyE(t, terraformOptions)
	if err == nil {
		return output, nil
	}

	var fatalErr retry.FatalError
	var cmdErr *shell.ErrWithCmdOutput
	if output == "" && errors.As(err, &fatalErr) && errors.As(fatalErr.Underlying, &cmdErr) {
		output = cmdErr.Output.Combined()
	}

	failures := Failures(output)
	if len(failures) == 0 {
		return output, err
	}

	// Some callers ignore apply errors when cleanup is disabled, so surface failures here too.
	for _, failure := range failures {
		logrus.Error(failure)
	}

	return output, fmt.Errorf("%s\n%w", strings.Join(failures, "\n"), err)
}

// Failures returns a one-line summary for each unique Terraform error block in the output.
func Failures(output string) []string {
	lines := strings.Split(output, "\n")

	history := map[string][]string{}
	for _, line := range lines {
		if resource, text, found := strings.Cut(line, remoteExecPrefix); found {
			resource = strings.TrimSpace(resource)
			history[resource] = append(history[resource], strings.TrimRight(text, " \r"))
		}
	}

	var failures []string
	seen := map[string]bool{}
	for i, line := range lines {
		title, ok := strings.CutPrefix(strings.TrimRight(line, " \r"), errorPrefix)
		if !ok || strings.Contains(line, remoteExecPrefix) {
			continue
		}

		resource, status, detail := errorBlockDetails(lines[i+1:])

		var summary string
		if status != "" {
			summary = resource + " script failed (exit status " + status + ")"
			if message := errorText(history[resource]); message != "" {
				summary += ": " + message
			}
		} else {
			summary = title
			if resource != "" {
				summary = resource + ": " + summary
			}
			if detail != "" {
				summary += ": " + detail
			}
		}

		if !seen[summary] {
			seen[summary] = true
			failures = append(failures, summary)
		}
	}

	return failures
}

// errorBlockDetails reads the resource details and script exit status from the lines following an error.
func errorBlockDetails(lines []string) (resource, status, detail string) {
	var details []string
	for _, line := range lines {
		if strings.HasPrefix(line, errorPrefix) || strings.HasPrefix(line, warningPrefix) || strings.Contains(line, remoteExecPrefix) {
			break
		}

		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
		case strings.HasPrefix(trimmed, "with "):
			if resource == "" {
				resource = strings.TrimSuffix(strings.TrimPrefix(trimmed, "with "), ",")
			}
		case isSourceContext(line, trimmed):
		default:
			if _, exit, ok := strings.Cut(trimmed, processExited); ok {
				status = exit
			}

			if len(details) < maxErrorLines {
				details = append(details, trimmed)
			}
		}
	}

	return resource, status, strings.Join(details, " ")
}

// isSourceContext reports whether a line is part of the config location and snippet Terraform prints with an error.
func isSourceContext(line, trimmed string) bool {
	if strings.HasPrefix(trimmed, "on ") && strings.HasPrefix(line, "  ") {
		return true
	}

	if strings.HasPrefix(trimmed, "├") || strings.HasPrefix(trimmed, "│") {
		return true
	}

	number, _, found := strings.Cut(trimmed, ":")
	if !found || number == "" || !strings.HasPrefix(line, " ") {
		return false
	}

	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// errorText skips trailing trace lines (e.g. "+ exit 1"), then collects output up to the previous trace line.
func errorText(lines []string) string {
	i := len(lines) - 1
	for i >= 0 && (isTrace(lines[i]) || strings.TrimSpace(lines[i]) == "") {
		i--
	}

	var message []string
	for ; i >= 0 && !isTrace(lines[i]) && len(message) < maxErrorLines; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			message = append([]string{trimmed}, message...)
		}
	}

	return strings.Join(message, " ")
}

func isTrace(line string) bool {
	return strings.HasPrefix(line, "+")
}
