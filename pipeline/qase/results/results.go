package results

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// ReportTest is a function that reports the test results.
func ReportTest(terratestConfig *config.TerratestConfig) error {
	runID := os.Getenv("QASE_TEST_RUN_ID")
	if runID == "" {
		logrus.Error("QASE_TEST_RUN_ID is not set")
		return nil
	}

	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	reporterPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/pipeline/scripts/build_qase_reporter.sh")

	cmd := exec.Command(reporterPath)
	output, err := cmd.Output()
	if err != nil {
		logrus.Error("Error running reporter script: ", err)
		return err
	}

	logrus.Info(string(output))

	return nil
}
