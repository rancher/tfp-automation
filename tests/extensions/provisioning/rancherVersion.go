package provisioning

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rancher/shepherd/extensions/rancherversion"
)

const (
	head = "head"
)

type Commit struct {
	SHA string `json:"sha"`
}

// RequestRancherVersion Requests the rancher version from the rancher server, parses the returned json and returns a
// Config object, or an error.
func RequestRancherVersion(rancherURL string) (*rancherversion.Config, error) {
	httpURL := "https://" + rancherURL + "/rancherversion"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(httpURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteObject, err := io.ReadAll(resp.Body)
	if err != nil || byteObject == nil {
		return nil, err
	}

	var jsonObject map[string]interface{}
	err = json.Unmarshal(byteObject, &jsonObject)
	if err != nil {
		return nil, err
	}

	configObject := new(rancherversion.Config)
	configObject.IsPrime, _ = strconv.ParseBool(jsonObject["RancherPrime"].(string))
	configObject.RancherVersion = jsonObject["Version"].(string)
	configObject.GitCommit = jsonObject["GitCommit"].(string)

	if strings.Contains(configObject.RancherVersion, head) {
		if i := strings.Index(configObject.RancherVersion, "-"); i != -1 {
			configObject.RancherVersion = configObject.RancherVersion[:i] + "-" + head
		}
	}

	return configObject, nil
}
