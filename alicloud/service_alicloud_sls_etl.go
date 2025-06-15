package alicloud

import (
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)
// CreateETL encapsulates the call to aliyunSlsAPI.CreateETL
func (s *SlsService) CreateETL(project string, etl *aliyunSlsAPI.ETL) error {
	return s.aliyunSlsAPI.CreateETL(project, etl)
}

// StartETL encapsulates the call to aliyunSlsAPI.StartETL
func (s *SlsService) StartETL(project, etlName string) error {
	return s.aliyunSlsAPI.StartETL(project, etlName)
}

// StopETL encapsulates the call to aliyunSlsAPI.StopETL
func (s *SlsService) StopETL(project, etlName string) error {
	return s.aliyunSlsAPI.StopETL(project, etlName)
}

// UpdateETL encapsulates the call to aliyunSlsAPI.UpdateETL
func (s *SlsService) UpdateETL(project, etlName string, etl *aliyunSlsAPI.ETL) error {
	return s.aliyunSlsAPI.UpdateETL(project, etlName, etl)
}

// DeleteETL encapsulates the call to aliyunSlsAPI.DeleteETL
func (s *SlsService) DeleteETL(project, etlName string) error {
	return s.aliyunSlsAPI.DeleteETL(project, etlName)
}
