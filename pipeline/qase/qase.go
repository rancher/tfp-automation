package qase

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/antihax/optional"
	qasedefaults "github.com/rancher/tests/actions/qase"
	"github.com/sirupsen/logrus"
	upstream "go.qase.io/client"
)

type TestSuiteSchema struct {
	Projects []string                  `json:"projects,omitempty" yaml:"projects,omitempty"`
	Suite    string                    `json:"suite,omitempty" yaml:"suite,omitempty"`
	Cases    []upstream.TestCaseCreate `json:"cases,omitempty" yaml:"cases,omitempty"`
}

type Service struct {
	Client *upstream.APIClient
}

const (
	schemas      = "schemas.yaml"
	requestLimit = 100
)

// SetupQaseClient creates a new Qase client from the api token environment variable QASE_AUTOMATION_TOKEN
func SetupQaseClient() *Service {
	cfg := upstream.NewConfiguration()
	cfg.AddDefaultHeader("Token", os.Getenv(qasedefaults.QaseTokenEnvVar))
	return &Service{
		Client: upstream.NewAPIClient(cfg),
	}
}

// GetTestSuite retrieves a Test Suite by name within a specified Qase Project if it exists
func (q *Service) GetTestSuite(project, suite string, parentID *int64) (*upstream.Suite, error) {
	logrus.Debugf("Getting test suite \"%s\" in project %s\n", suite, project)

	var numOfSuites int32 = 1
	offSetCount := 0
	for numOfSuites > 0 {
		localVarOptionals := &upstream.SuitesApiGetSuitesOpts{
			Limit:         optional.NewInt32(int32(requestLimit)),
			Offset:        optional.NewInt32(int32(offSetCount)),
			FiltersSearch: optional.NewString(suite),
		}
		qaseSuites, _, err := q.Client.SuitesApi.GetSuites(context.TODO(), project, localVarOptionals)
		if err != nil {
			logrus.Info(err)
			return nil, err
		}

		numOfSuites = qaseSuites.Result.Count
		for _, result := range qaseSuites.Result.Entities {
			if result.ParentId == *parentID && result.Title == suite {
				return &result, nil
			}
		}

		offSetCount += requestLimit
	}

	return nil, fmt.Errorf("test suite \"%s\" not found in project %s", suite, project)
}

// CreateTestSuite creates a new Test Suite within a specified Qase Project
func (q *Service) CreateTestSuite(project string, suite upstream.SuiteCreate) (int64, error) {
	logrus.Debugf("Creating test suite \"%s\" in project %s\n", suite.Title, project)
	resp, _, err := q.Client.SuitesApi.CreateSuite(context.TODO(), suite, project)
	if err != nil {
		return 0, fmt.Errorf("failed to create test suite: \"%s\". Error: %v", suite.Title, err)
	}
	return resp.Result.Id, nil
}

// createTestCase creates a new test in qase
func (q *Service) createTestCase(project string, testCase upstream.TestCaseCreate) error {
	_, _, err := q.Client.CasesApi.CreateCase(context.TODO(), testCase, project)
	if err != nil {
		return fmt.Errorf("failed to create test case: \"%s\". Error: %v", testCase.Title, err)
	}
	return nil
}

// updateTestCase updates an existing test in qase
func (q *Service) updateTestCase(project string, testCase upstream.TestCaseUpdate, id int32) error {
	_, _, err := q.Client.CasesApi.UpdateCase(context.TODO(), testCase, project, id)
	if err != nil {
		return fmt.Errorf("failed to update test case: \"%s\". Error: %v", testCase.Title, err)
	}
	return nil
}

// createSuitePath creates a series of nested test suites from a / deliniated string
func createSuitePath(client *Service, suiteName, project string) (int64, error) {
	suites := strings.Split(suiteName, "/")
	testSuiteId := int64(0)
	var parentID *int64
	for _, suite := range suites {
		parentID = &testSuiteId

		testSuite, err := client.GetTestSuite(project, suite, parentID)
		if testSuite != nil {
			testSuiteId = testSuite.Id
		}

		if err != nil && testSuite != nil {
			logrus.Error("Could not determine test suite:", err)
			return 0, err
		} else if err != nil {
			logrus.Debug("Error obtaining test suite:", err)
			suiteBody := upstream.SuiteCreate{Title: suite}
			if testSuiteId != 0 {
				suiteBody.ParentId = testSuiteId
			}
			testSuiteId, _ = client.CreateTestSuite(project, suiteBody)
		}
	}

	return testSuiteId, nil
}

// UploadTests either creates new Test Cases and their associated Suite or updates them if they already exist
func (q *Service) UploadTests(project string, testCases []upstream.TestCaseCreate) error {
	for _, tc := range testCases {
		existingCase, err := q.getTestCase(project, tc)
		if err == nil {
			logrus.Info("Updating test case:\n\tProject: ", project, "\n\tTitle: ", tc.Title, "\n\tSuiteId: ", tc.SuiteId)
			var qaseTest upstream.TestCaseUpdate
			qaseTest.Title = tc.Title
			qaseTest.SuiteId = tc.SuiteId
			qaseTest.Description = tc.Description
			qaseTest.Type_ = tc.Type_
			qaseTest.Priority = tc.Priority
			qaseTest.IsFlaky = tc.IsFlaky
			qaseTest.Automation = tc.Automation
			qaseTest.Params = tc.Params
			qaseTest.CustomField = tc.CustomField
			for _, step := range tc.Steps {
				var qaseSteps upstream.TestCaseUpdateSteps
				qaseSteps.Action = step.Action
				qaseSteps.ExpectedResult = step.ExpectedResult
				qaseSteps.Data = step.Data
				qaseSteps.Position = step.Position
				qaseTest.Steps = append(qaseTest.Steps, qaseSteps)
			}
			err = q.updateTestCase(project, qaseTest, int32(existingCase.Id))
			if err != nil {
				return err
			}
		} else if existingCase != nil {
			return err
		} else {
			logrus.Info("Uploading test case:\n\tProject: ", project, "\n\tTitle: ", tc.Title, "\n\tDescription: ", tc.Description, "\n\tSuiteId: ", tc.SuiteId, "\n\tAutomation: ", tc.Automation, "\n\tSteps: ", tc.Steps)
			err = q.createTestCase(project, tc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getTestCase retrieves a Test Case by name within a specified Qase Project if it exists
func (q *Service) getTestCase(project string, test upstream.TestCaseCreate) (*upstream.TestCase, error) {
	logrus.Debugf("Getting test case \"%s\" in project %s\n", test.Title, project)
	localVarOptionals := &upstream.CasesApiGetCasesOpts{
		FiltersSearch:  optional.NewString(test.Title),
		FiltersSuiteId: optional.NewInt32(int32(test.SuiteId)),
	}
	qaseTestCases, _, err := q.Client.CasesApi.GetCases(context.TODO(), project, localVarOptionals)
	if err != nil {
		return nil, err
	}

	resultLength := len(qaseTestCases.Result.Entities)
	if resultLength == 1 {
		return &qaseTestCases.Result.Entities[0], nil
	} else if resultLength > 1 {
		for _, entity := range qaseTestCases.Result.Entities {
			if entity.Title == test.Title {
				return &entity, nil
			}
		}

		return &qaseTestCases.Result.Entities[0], fmt.Errorf("test case \"%s\" found multiple times in project %s, but should only exist once", test.Title, project)
	}

	return nil, fmt.Errorf("test case \"%s\" not found in project %s", test.Title, project)
}
