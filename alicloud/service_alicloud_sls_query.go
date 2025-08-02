package alicloud

import (
	"fmt"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

// QuerySlsLogs executes a log query and returns the results
func (s *SlsService) QuerySlsLogs(projectName, logstoreName string, from, to int32, query string, lineNum int64) (*sls.LogResult, error) {
	slsAPI := s.GetAPI()

	// Validate required parameters
	if projectName == "" {
		return nil, WrapError(fmt.Errorf("project name is required"))
	}
	if logstoreName == "" {
		return nil, WrapError(fmt.Errorf("logstore name is required"))
	}

	// Execute simple log query using SLS API
	result, err := slsAPI.QueryLogSimple(projectName, logstoreName, from, to, query, lineNum)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}
