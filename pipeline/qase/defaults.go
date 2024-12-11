package qase

const (
	AutomationSuiteID       = int32(554)
	AutomationTestNameID    = 15
	FailStatus              = "fail"
	MultiSubTestPattern     = `(\w+/\w+/\w+){1,}`
	PassStatus              = "pass"
	QaseTokenEnvVar         = "QASE_AUTOMATION_TOKEN"
	RancherManagerProjectID = "RM"
	RecurringRunID          = 1
	RunSourceID             = 16
	SkipStatus              = "skip"
	SubtestPattern          = `(\w+/\w+){1,1}`
	TestResultsJSON         = "results.json"
	TestRunEnvVar           = "QASE_TEST_RUN_ID"
	TestRunNAMEEnvVar       = "TEST_RUN_NAME"
	TestSource              = "GoValidation"
	TestSourceID            = 14
)
