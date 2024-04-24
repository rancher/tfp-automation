package testrun

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	defaults "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/sirupsen/logrus"
	qase "go.qase.io/client"
	"gopkg.in/yaml.v2"
)

type RecurringTestRun struct {
	ID int64 `json:"id" yaml:"id"`
}

func main() {
	// commandline flags
	startRun := flag.Bool("startRun", false, "commandline flag that determines when to start a run, and conversely when to end it.")
	flag.Parse()

	qaseToken := os.Getenv(defaults.QaseTokenEnvVar)

	cfg := qase.NewConfiguration()
	cfg.AddDefaultHeader("Token", qaseToken)
	client := qase.NewAPIClient(cfg)

	if *startRun {
		// create test run
		runIDEnvVar := os.Getenv(defaults.TestRunNAMEEnvVar)

		resp, err := createTestRun(client, runIDEnvVar)
		if err != nil {
			logrus.Error("error creating test run: ", err)
		}

		newRunID := resp.Result.Id
		recurringTestRun := RecurringTestRun{}
		recurringTestRun.ID = newRunID
		err = writeToConfigFile(recurringTestRun)
		if err != nil {
			logrus.Error("error writiing test run config: ", err)
		}
	} else {

		testRunConfig, err := readConfigFile()
		if err != nil {
			logrus.Fatalf("error reporting converting string to int32: %v", err)
		}
		// complete test run
		_, _, err = client.RunsApi.CompleteRun(context.TODO(), defaults.RancherManagerProjectID, int32(testRunConfig.ID))
		if err != nil {
			log.Fatalf("error completing test run: %v", err)
		}
	}

}

func createTestRun(client *qase.APIClient, testRunName string) (*qase.IdResponse, error) {
	runCreateBody := qase.RunCreate{
		Title: testRunName,
		CustomField: map[string]string{
			fmt.Sprintf("%d", defaults.RunSourceID): fmt.Sprintf("%d", defaults.RecurringRunID),
		},
	}

	idResponse, _, err := client.RunsApi.CreateRun(context.TODO(), runCreateBody, defaults.RancherManagerProjectID)
	if err != nil {
		return nil, err
	}

	return &idResponse, nil
}

func writeToConfigFile(config RecurringTestRun) error {
	yamlConfig, err := yaml.Marshal(config)

	if err != nil {
		return err
	}

	return os.WriteFile("testrunconfig.yaml", yamlConfig, 0644)
}

func readConfigFile() (*RecurringTestRun, error) {
	configString, err := os.ReadFile("testrunconfig.yaml")
	if err != nil {
		return nil, err
	}

	var testRunConfig RecurringTestRun
	err = yaml.Unmarshal(configString, &testRunConfig)
	if err != nil {
		return nil, err
	}

	return &testRunConfig, nil
}
