package alicloud

// StartInstanceRequest represents the request to start a Kafka instance
type StartInstanceRequest struct {
	InstanceId           string
	RegionId             string
	VpcId                string
	VSwitchId            string
	ZoneId               string
	DeployModule         string
	IsEipInner           bool
	IsSetUserAndPassword bool
	Username             string
	Password             string
	Name                 string
	CrossZone            bool
	SecurityGroup        string
	ServiceVersion       string
	Config               string
	KMSKeyId             string
	Notifier             string
	UserPhoneNum         string
	SelectedZones        string
	IsForceSelectedZones bool
	VSwitchIds           []string
}

// ModifyInstanceNameRequest represents the request to modify a Kafka instance name
type ModifyInstanceNameRequest struct {
	InstanceId   string
	RegionId     string
	InstanceName string
}

// UpgradeInstanceVersionRequest represents the request to upgrade a Kafka instance version
type UpgradeInstanceVersionRequest struct {
	InstanceId    string
	RegionId      string
	TargetVersion string
}

// StopInstanceRequest represents the request to stop a Kafka instance
type StopInstanceRequest struct {
	InstanceId string
	RegionId   string
}
