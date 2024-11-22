package aws

type Config struct {
	AMI                   string   `json:"ami,omitempty" yaml:"ami,omitempty"`
	AWSInstanceType       string   `json:"awsInstanceType,omitempty" yaml:"awsInstanceType,omitempty"`
	AWSKeyName            string   `json:"awsKeyName,omitempty" yaml:"awsKeyName,omitempty"`
	AWSRootSize           int64    `json:"awsRootSize,omitempty" yaml:"awsRootSize,omitempty"`
	AWSSecurityGroupNames []string `json:"awsSecurityGroupNames,omitempty" yaml:"awsSecurityGroupNames,omitempty"`
	AWSSecurityGroups     []string `json:"awsSecurityGroups,omitempty" yaml:"awsSecurityGroups,omitempty"`
	AWSSubnetID           string   `json:"awsSubnetID,omitempty" yaml:"awsSubnetID,omitempty"`
	AWSSubnets            []string `json:"awsSubnets,omitempty" yaml:"awsSubnets,omitempty"`
	AWSVpcID              string   `json:"awsVpcID,omitempty" yaml:"awsVpcID,omitempty"`
	AWSRoute53Zone        string   `json:"awsRoute53Zone,omitempty" yaml:"awsRoute53Zone,omitempty"`
	AWSZoneLetter         string   `json:"awsZoneLetter,omitempty" yaml:"awsZoneLetter,omitempty"`
	PrivateAccess         bool     `json:"privateAccess,omitempty" yaml:"privateAccess,omitempty"`
	PublicAccess          bool     `json:"publicAccess,omitempty" yaml:"publicAccess,omitempty"`
	Region                string   `json:"region,omitempty" yaml:"region,omitempty"`
	AWSUser               string   `json:"awsUser,omitempty" yaml:"awsUser,omitempty"`
	Timeout               string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}
