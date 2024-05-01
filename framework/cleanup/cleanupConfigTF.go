package cleanup

import (
	"os"

	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/sirupsen/logrus"
)

// CleanupConfigTF is a function that will cleanup the main.tf file and terraform.tfstate files.
func CleanupConfigTF() error {
	keyPath := resources.SetKeyPath()

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Errorf("Failed to overwrite main.tf file. Error: %v", err)

		return err
	}

	defer file.Close()

	_, err = file.WriteString("// Leave blank - main.tf will be set during testing")
	if err != nil {
		logrus.Errorf("Failed to write to main.tf file. Error: %v", err)

		return err
	}

	delete_files := [3]string{configs.TFState, configs.TFStateBackup, configs.TFLockHCL}

	for _, delete_file := range delete_files {
		delete_file = keyPath + delete_file
		err = os.Remove(delete_file)

		if err != nil {
			logrus.Errorf("Failed to delete terraform.tfstate, terraform.tfstate.backup, and terraform.lock.hcl files. Error: %v", err)

			return err
		}
	}

	err = os.RemoveAll(keyPath + configs.TerraformFolder)
	if err != nil {
		logrus.Errorf("Failed to delete .terraform folder. Error: %v", err)

		return err
	}

	return nil
}
