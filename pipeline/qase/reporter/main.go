package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/rancher/tests/actions/qase"
	qaseactions "github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/qase/testresult"
	"github.com/sirupsen/logrus"
	upstream "go.qase.io/qase-api-client"
	"gopkg.in/yaml.v2"
)

var (
	runIDEnvVar             = os.Getenv(qase.TestRunEnvVar)
	projectIDEnvVar         = os.Getenv(qase.ProjectIDEnvVar)
	testRunName             = os.Getenv(qase.TestRunNameEnvVar)
	_, callerFilePath, _, _ = runtime.Caller(0)
	basepath                = filepath.Join(filepath.Dir(callerFilePath), "..", "..", "..")
	validStatus             = []string{"passed", "failed", "skipped"}
)

const (
	requestLimit = 100
)

func main() {
	logrus.Info("Running QASE reporter")
	if projectIDEnvVar == "" {
		logrus.Warningf("Project env var not provided, defaulting to %s", qaseactions.RancherManagerProjectID)
		projectIDEnvVar = qaseactions.RancherManagerProjectID
	}

	if runIDEnvVar != "" {
		client := qase.SetupQaseClient()

		runID, err := strconv.ParseInt(runIDEnvVar, 10, 64)

		if testRunName != "" {
			resp, err := client.CreateTestRun(testRunName, projectIDEnvVar)
			if err != nil {
				logrus.Error("error creating test run: ", err)
			} else {
				runID = *resp.Result.Id
			}
		}

		if err != nil {
			logrus.Fatalf("error reporting converting string to int64: %v", err)
		}

		err = reportTestQases(client, int32(runID))
		if err != nil {
			logrus.Error("error reporting: ", err)
		}
	} else {
		logrus.Warningf("QASE run ID not provided")
	}
}

// getAllAutomationTestCases gets all qase tests in a project
func getAllAutomationTestCases(qaseService *qase.Service) (map[string]upstream.TestCase, error) {
	testCases := []upstream.TestCase{}
	testCaseNameMap := map[string]upstream.TestCase{}
	var numOfTestsCases int32 = 1
	offSetCount := 0
	for numOfTestsCases > 0 {
		casesRequest := qaseService.Client.CasesAPI.GetCases(context.TODO(), projectIDEnvVar)

		casesRequest = casesRequest.Offset(int32(offSetCount))
		casesRequest = casesRequest.Limit(requestLimit)
		resp, _, _ := casesRequest.Execute()

		testCases = append(testCases, resp.Result.Entities...)
		numOfTestsCases = *resp.Result.Count
		offSetCount += len(resp.Result.Entities)
	}

	for _, testCase := range testCases {
		automationTestNameCustomField := getAutomationTestName(testCase.CustomFields)
		if automationTestNameCustomField != "" {
			testCaseNameMap[automationTestNameCustomField] = testCase
		} else {
			testCaseNameMap[*testCase.Title] = testCase
		}

	}

	return testCaseNameMap, nil
}

// readTestResults converts the results.json file into an output object
func readTestResults() ([]testresult.GoTestOutput, error) {
	file, err := os.Open(qase.TestResultsJSON)
	if err != nil {
		return nil, err
	}

	fscanner := bufio.NewScanner(file)
	outputLines := []testresult.GoTestOutput{}
	for fscanner.Scan() {
		var testCase testresult.GoTestOutput
		err = yaml.Unmarshal(fscanner.Bytes(), &testCase)
		if err != nil {
			return nil, err
		}
		outputLines = append(outputLines, testCase)
	}

	return outputLines, nil
}

// parseTestResults parses the results.json into a test results object
func parseTestResults(outputs []testresult.GoTestOutput) map[string]*testresult.GoTestResult {
	finalTestResults := map[string]*testresult.GoTestResult{}
	var timeoutFailure bool

	for _, output := range outputs {
		var tableTestName string
		testName := strings.Split(output.Test, "/")
		if len(testName) > 1 {
			tableTestName = testName[len(testName)-1]
		}

		if output.Action == "run" && tableTestName != "" {
			newTestResult := &testresult.GoTestResult{Name: tableTestName, Package: output.Package}
			finalTestResults[tableTestName] = newTestResult
		} else if output.Action == "output" && tableTestName != "" {
			goTestResult := finalTestResults[tableTestName]
			goTestResult.StackTrace += output.Output
		} else if output.Action == qase.SkipStatus {
			if tableTestName != "" {
				delete(finalTestResults, tableTestName)
			}
		} else if (output.Action == qase.FailStatus || output.Action == qase.PassStatus) && tableTestName != "" {
			if tableTestName != "" {
				goTestResult := finalTestResults[tableTestName]
				goTestResult.StackTrace += output.Output
				goTestResult.Status = output.Action
				goTestResult.Elapsed = output.Elapsed
			} else {
				goTestResult := finalTestResults[tableTestName]
				goTestResult.StackTrace += output.Output
			}
		} else if output.Action == qase.FailStatus && tableTestName != "" {
			timeoutFailure = true
		}
	}

	for _, testResult := range finalTestResults {
		testSuite := strings.Split(testResult.Name, "/")
		testName := testSuite[len(testSuite)-1]
		testResult.Name = testName
		testResult.TestSuite = testSuite[0 : len(testSuite)-1]
		if timeoutFailure && testResult.Status == "" {
			testResult.Status = qase.FailStatus
		}
	}

	return finalTestResults
}

// reportTestQases updates a qase test run with the results of a set of tests
func reportTestQases(qaseService *qase.Service, testRunID int32) error {
	resultsOutputs, err := readTestResults()
	if err != nil {
		return nil
	}

	goTestResults := parseTestResults(resultsOutputs)

	qaseTestCases, err := getAllAutomationTestCases(qaseService)
	if err != nil {
		return err
	}

	for _, goTestResult := range goTestResults {
		if testQase, ok := qaseTestCases[goTestResult.Name]; ok {
			basePathDirs := strings.Split(basepath, "/")
			baseTestPathDir := basePathDirs[len(basePathDirs)-1]

			packagePath := strings.Split(goTestResult.Package, baseTestPathDir)
			if len(packagePath) > 2 {
				return errors.New("Error base path directory is not unique")
			}

			fullPackagePath := filepath.Join(basepath, packagePath[1])
			qaseProjects, err := qase.GetSchemas(fullPackagePath)
			if err != nil {
				logrus.Warning(err)
				continue
			}

			qaseTestSchema, err := qase.GetTestSchema(goTestResult.Name, qaseProjects)
			if err != nil {
				logrus.Warning(err)
				continue
			}

			// update test status
			logrus.Infof("Updating run with %v", *testQase.Title)
			err = updateTestInRun(qaseService.Client, *goTestResult, testQase, qaseTestSchema.Parameters, testRunID)
			if err != nil {
				logrus.Warning(err)
				continue
			}
		} else {
			err = fmt.Errorf("Test case not found in qase: %s", goTestResult.Name)
			logrus.Warning(err)
			continue
		}
	}

	return nil
}

// updateTestInRun updates the current qase test run with a test
func updateTestInRun(client *upstream.APIClient, testResult testresult.GoTestResult, qaseTestCase upstream.TestCase, params []upstream.TestCaseParameterCreate, testRunID int32) error {
	var elapsedTime int64
	var err error
	if testResult.Elapsed != "" {
		floatTime, err := strconv.ParseFloat(testResult.Elapsed, 64)
		if err != nil {
			return err
		}
		elapsedTime = int64(floatTime)
	}

	resultParams := make(map[string]string)
	for _, param := range params {
		if param.ParameterSingle == nil {
			continue
		}

		if len(param.ParameterSingle.Values) > 0 {
			paramKey := param.ParameterSingle.Title
			paramVal := strings.Join(param.ParameterSingle.Values, ", ")
			if paramVal == "" {
				continue
			}

			resultParams[paramKey] = paramVal
		}
	}

	status := fmt.Sprintf("%sed", testResult.Status)
	if !slices.Contains(validStatus, status) {
		status = "failed"
	}

	resultBody := upstream.ResultCreate{
		CaseId:  qaseTestCase.Id,
		Status:  status,
		Time:    *upstream.NewNullableInt64(&elapsedTime),
		Param:   resultParams,
		Comment: *upstream.NewNullableString(&testResult.StackTrace),
	}

	resultRequest := client.ResultsAPI.CreateResult(context.TODO(), projectIDEnvVar, testRunID)
	resultRequest = resultRequest.ResultCreate(resultBody)

	_, _, err = resultRequest.Execute()
	if err != nil {
		return err
	}

	return nil
}

// getAutomationTestName gets the custom test name field
func getAutomationTestName(customFields []upstream.CustomFieldValue) string {
	for _, field := range customFields {
		if *field.Id == qase.AutomationTestNameID {
			return *field.Value
		}
	}
	return ""
}
