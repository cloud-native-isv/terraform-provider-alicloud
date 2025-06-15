package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// ======== Legacy Compatibility Functions ========
// These functions provide backward compatibility for existing resources
// that still use the old LogService interface

// Legacy LogService struct for backward compatibility
type LogService struct {
	client *connectivity.AliyunClient
	sls    *SlsService
}

// NewLogService creates a new LogService instance that wraps SlsService
func NewLogService(client *connectivity.AliyunClient) *LogService {
	slsService, _ := NewSlsService(client)
	return &LogService{
		client: client,
		sls:    slsService,
	}
}

// Legacy function wrappers that delegate to SlsService
func (s *LogService) DescribeLogProject(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsProject(id)
}

func (s *LogService) DescribeLogStore(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsLogStore(id)
}

func (s *LogService) DescribeLogStoreIndex(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsLogStoreIndex(id)
}

func (s *LogService) DescribeLogMachineGroup(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsMachineGroup(id)
}

func (s *LogService) DescribeLogtailConfig(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsLogtailConfig(id)
}

func (s *LogService) DescribeLogtailAttachment(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsLogtailAttachment(id)
}

func (s *LogService) DescribeLogAlert(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsAlert(id)
}

func (s *LogService) DescribeLogAlertResource(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsLogAlertResource(id)
}

func (s *LogService) DescribeLogDashboard(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsDashboard(id)
}

func (s *LogService) DescribeLogEtl(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsEtl(id)
}

func (s *LogService) DescribeLogIngestion(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsIngestion(id)
}

func (s *LogService) DescribeLogOssExport(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsOssExportSink(id)
}

func (s *LogService) DescribeLogOssShipper(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsOssShipper(id)
}

func (s *LogService) DescribeLogResource(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsResource(id)
}

func (s *LogService) DescribeLogResourceRecord(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsResourceRecord(id)
}

func (s *LogService) DescribeLogAudit(id string) (map[string]interface{}, error) {
	return s.sls.DescribeSlsAudit(id)
}

func (s *LogService) DescribeLogProjectTags(projectName string) (map[string]interface{}, error) {
	return s.sls.DescribeListTagResources(projectName)
}

// Legacy Wait functions that use SlsService StateRefreshFunc
func (s *LogService) WaitForLogProject(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"Normal"},
		Refresh:    s.sls.SlsProjectStateRefreshFunc(id, "status", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"Normal"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogStore(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsLogStoreStateRefreshFunc(id, "#logstoreName", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogMachineGroup(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsMachineGroupStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogtailConfig(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsLogtailConfigStateRefreshFunc(id, "#configName", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogtailAttachment(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"true"},
		Refresh:    s.sls.SlsLogtailAttachmentStateRefreshFunc(id, "attached", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"true"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogstoreAlert(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsAlertStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogDashboard(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsDashboardStateRefreshFunc(id, "#dashboardName", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogETL(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsEtlStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogIngestion(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsIngestionStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogOssExport(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsOssExportSinkStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogOssShipper(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsOssShipperStateRefreshFunc(id, "#shipperName", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogResource(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsResourceStateRefreshFunc(id, "#name", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

func (s *LogService) WaitForLogResourceRecord(id string, status Status, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{"#CHECKSET"},
		Refresh:    s.sls.SlsResourceRecordStateRefreshFunc(id, "#id", []string{}),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if status == Deleted {
		stateConf.Target = []string{""}
		stateConf.Pending = []string{"#CHECKSET"}
	}

	_, err := stateConf.WaitForState()
	return WrapError(err)
}

// Missing StateRefreshFunc for SlsOssShipper
func (s *SlsService) SlsOssShipperStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsOssShipper(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		if strings.HasPrefix(field, "#") {
			v, _ := jsonpath.Get(strings.TrimPrefix(field, "#"), object)
			if v != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}
