package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateNasMountTarget creates a NAS mount target using CWS-Lib-Go API
func (s *NasService) CreateNasMountTarget(fileSystemId string, mountTarget *aliyunNasAPI.MountTarget) (*aliyunNasAPI.MountTarget, error) {
	// Get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// Create mount target
	createdMountTarget, err := nasAPI.CreateMountTarget(fileSystemId, mountTarget)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_mount_target", "CreateMountTarget", AlibabaCloudSdkGoERROR)
	}

	return createdMountTarget, nil
}

// DeleteNasMountTarget deletes a NAS mount target using CWS-Lib-Go API
func (s *NasService) DeleteNasMountTarget(fileSystemId, mountTargetDomain string) error {
	// Get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// Delete mount target
	err = nasAPI.DeleteMountTarget(fileSystemId, mountTargetDomain)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", fileSystemId, mountTargetDomain), "DeleteMountTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// GetNasMountTarget gets a single NAS mount target using CWS-Lib-Go API
func (s *NasService) DescribeNasMountTarget(id string) (*aliyunNasAPI.MountTarget, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	fileSystemId := parts[0]
	mountTargetDomain := parts[1]

	// Get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// List all mount targets for the file system
	mountTargets, err := nasAPI.ListMountTargets(fileSystemId, mountTargetDomain)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidMountTarget.NotFound", "InvalidFileSystem.NotFound", "Forbidden.NasNotFound", "InvalidLBid.NotFound", "VolumeUnavailable", "InvalidParam.MountTargetDomain"}) {
			return nil, WrapErrorf(NotFoundErr("MountTarget", id), NotFoundMsg, err)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "ListMountTargets", AlibabaCloudSdkGoERROR)
	}

	// Find the specific mount target by domain
	for _, mt := range mountTargets {
		if mt.MountTargetDomain == mountTargetDomain {
			return &mt, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("MountTarget", id), NotFoundMsg, "MountTarget not found")
}

// ListNasMountTargets lists all mount targets for a file system using CWS-Lib-Go API
func (s *NasService) ListNasMountTargets(fileSystemId string) ([]aliyunNasAPI.MountTarget, error) {
	// Get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// List all mount targets for the file system
	mountTargets, err := nasAPI.ListMountTargets(fileSystemId, "")
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ListMountTargets", AlibabaCloudSdkGoERROR)
	}

	return mountTargets, nil
}

// UpdateNasMountTarget updates a NAS mount target using CWS-Lib-Go API
func (s *NasService) UpdateNasMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status string) error {
	// Get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// Update mount target
	err = nasAPI.ModifyMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", fileSystemId, mountTargetDomain), "ModifyMountTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// NasMountTargetStateRefreshFunc returns a StateRefreshFunc for NAS mount target status
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
	stateConf := BuildStateConf([]string{"Creating", "Pending"}, []string{status}, time.Duration(timeout)*time.Second, 5*time.Second, s.NasMountTargetStateRefreshFunc(id, "$.Status", []string{"Failed", "Error"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
