package main

import (
	"path/filepath"
	"runtime"

	"github.com/rancher/tests/actions/qase"
	"github.com/sirupsen/logrus"
)

var (
	_, callerFilePath, _, _ = runtime.Caller(0)
	basepath                = filepath.Join(filepath.Dir(callerFilePath), "..", "..", "..", "..")
)

func main() {
	client := qase.SetupQaseClient()

	err := qase.UploadSchemas(client, basepath)
	if err != nil {
		logrus.Error(err)
	}
}
