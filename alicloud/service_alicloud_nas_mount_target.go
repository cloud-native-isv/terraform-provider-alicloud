package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *NasService) CreateNasMountTarget(fileSystemId string, mountTarget *aliyunNasAPI.MountTarget) (*aliyunNasAPI.MountTarget, error) {
	nasAPI := s.GetAPI()

	createdMountTarget, err := nasAPI.CreateMountTarget(fileSystemId, mountTarget)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_mount_target", "CreateMountTarget", AlibabaCloudSdkGoERROR)
	}

	return createdMountTarget, nil
}

func (s *NasService) DeleteNasMountTarget(fileSystemId, mountTargetDomain string) error {
	nasAPI := s.GetAPI()

	err := nasAPI.DeleteMountTarget(fileSystemId, mountTargetDomain)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", fileSystemId, mountTargetDomain), "DeleteMountTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *NasService) DescribeNasMountTarget(id string) (*aliyunNasAPI.MountTarget, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	fileSystemId := parts[0]
	mountTargetDomain := parts[1]

	nasAPI := s.GetAPI()

	mountTargets, err := nasAPI.ListMountTargets(fileSystemId, mountTargetDomain)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidMountTarget.NotFound", "InvalidFileSystem.NotFound", "Forbidden.NasNotFound", "InvalidLBid.NotFound", "VolumeUnavailable", "InvalidParam.MountTargetDomain"}) {
			return nil, WrapErrorf(NotFoundErr("MountTarget", id), NotFoundMsg, err)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "ListMountTargets", AlibabaCloudSdkGoERROR)
	}

	for _, mt := range mountTargets {
		if mt.MountTargetDomain == mountTargetDomain {
			return &mt, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("MountTarget", id), NotFoundMsg, "MountTarget not found")
}

func (s *NasService) ListNasMountTargets(fileSystemId string) ([]aliyunNasAPI.MountTarget, error) {
	nasAPI := s.GetAPI()

	mountTargets, err := nasAPI.ListMountTargets(fileSystemId, "")
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ListMountTargets", AlibabaCloudSdkGoERROR)
	}

	return mountTargets, nil
}

func (s *NasService) UpdateNasMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status string) error {
	nasAPI := s.GetAPI()

	err := nasAPI.ModifyMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", fileSystemId, mountTargetDomain), "ModifyMountTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *NasService) NasMountTargetStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		mountTarget, err := s.DescribeNasMountTarget(id)
		if err != nil {
			if NotFoundError(err) {
				return mountTarget, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string

		// Handle different field types for mount target status
		switch field {
		case "$.Status":
			currentStatus = mountTarget.Status
		case "#CHECKSET":
			// For checkset operations, if mount target exists, it's considered set
			if mountTarget != nil {
				currentStatus = "#CHECKSET"
			}
		default:
			// For other fields, use Status as default
			currentStatus = mountTarget.Status
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return mountTarget, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return mountTarget, currentStatus, nil
	}
}

// WaitForNasMountTarget waits for a mount target to reach the specified status
func (s *NasService) WaitForNasMountTarget(id string, status string, timeout int) error {
	// Define all possible intermediate states
	intermediateStates := []string{"Pending", "Creating", "Inactive", "Deleting", "Hibernating"}

	// Active, Hibernated and empty string are final states
	finalStates := []string{"Active", "Hibernated", ""}

	// Check if the requested status is a final state
	isFinalState := false
	for _, finalState := range finalStates {
		if status == finalState {
			isFinalState = true
			break
		}
	}

	if !isFinalState {
		// If status is not a final state, add it to intermediate states
		intermediateStates = append(intermediateStates, status)
	}

	// Final target states
	targetStates := []string{status}

	stateConf := BuildStateConf(intermediateStates, targetStates, time.Duration(timeout)*time.Second, 5*time.Second, s.NasMountTargetStateRefreshFunc(id, "$.Status", []string{"Failed", "Error"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForNasMountTargetCreated waits for a mount target to be created successfully
func (s *NasService) WaitForNasMountTargetCreated(id string, timeout int) error {
	stateConf := BuildStateConf([]string{"Pending", "Creating"}, []string{"Active"}, time.Duration(timeout)*time.Second, 5*time.Second, s.NasMountTargetStateRefreshFunc(id, "$.Status", []string{"Failed", "Error"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForNasMountTargetDeleted waits for a mount target to be deleted successfully
func (s *NasService) WaitForNasMountTargetDeleted(id string, timeout int) error {
	stateConf := BuildStateConf([]string{"Active", "Inactive", "Deleting"}, []string{""}, time.Duration(timeout)*time.Second, 5*time.Second, s.NasMountTargetStateRefreshFunc(id, "$.Status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
