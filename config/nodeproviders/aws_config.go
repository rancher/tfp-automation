package nodeproviders

type AWSConfig struct {
	AMI                   string   `json:"ami,omitempty" yaml:"ami,omitempty"`
	AWSAccessKey          string   `json:"awsAccessKey,omitempty" yaml:"awsAccessKey,omitempty"`
	AWSInstanceType       string   `json:"awsInstanceType,omitempty" yaml:"awsInstanceType,omitempty"`
	AWSRootSize           int64    `json:"awsRootSize,omitempty" yaml:"awsRootSiz,omitemptye"`
	AWSSecretKey          string   `json:"awsSecretKey,omitempty" yaml:"awsSecretKey,omitempty"`
	AWSSecurityGroupNames []string `json:"awsSecurityGroupNames,omitempty" yaml:"awsSecurityGroupNames,omitempty"`
	AWSSecurityGroups     []string `json:"awsSecurityGroups,omitempty" yaml:"awsSecurityGroups,omitempty"`
	AWSSubnetID           string   `json:"awsSubnetID,omitempty" yaml:"awsSubnetID,omitempty"`
	AWSSubnets            []string `json:"awsSubnets,omitempty" yaml:"awsSubnets,omitempty"`
	AWSVpcID              string   `json:"awsVpcID,omitempty" yaml:"awsVpcID,omitempty"`
	AWSZoneLetter         string   `json:"awsZoneLetter,omitempty" yaml:"awsZoneLetter,omitempty"`
	PrivateAccess         bool     `json:"privateAccess,omitempty" yaml:"privateAccess,omitempty"`
	PublicAccess          bool     `json:"publicAccess,omitempty" yaml:"publicAccess,omitempty"`
	Region                string   `json:"region,omitempty" yaml:"region,omitempty"`
}
