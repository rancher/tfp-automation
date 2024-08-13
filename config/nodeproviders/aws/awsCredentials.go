package aws

type Credentials struct {
	AWSAccessKey string `json:"awsAccessKey,omitempty" yaml:"awsAccessKey,omitempty"`
	AWSSecretKey string `json:"awsSecretKey,omitempty" yaml:"awsSecretKey,omitempty"`
}
