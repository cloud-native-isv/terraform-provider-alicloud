package alicloud

import (
	"fmt"
	"strings"
	"time"

	common "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeNasFileSystem gets NAS file system information using CWS-Lib-Go API
func (s *NasService) DescribeNasFileSystem(id string) (fileSystem *aliyunNasAPI.FileSystem, err error) {
	nasAPI := s.aliyunNasAPI

	fileSystem, err = nasAPI.GetFileSystem(id)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, WrapErrorf(NotFoundErr("FileSystem", id), NotFoundMsg, err)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetFileSystem", AlibabaCloudSdkGoERROR)
	}

	return fileSystem, nil
}

// NasFileSystemStateRefreshFunc returns a StateRefreshFunc for NAS file system status
func (s *NasService) NasFileSystemStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fileSystem, err := s.DescribeNasFileSystem(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string

		// Use strong typing to get field values instead of map conversion
		switch field {
		case "Status":
			currentStatus = fileSystem.Status
		case "FileSystemId":
			currentStatus = fileSystem.FileSystemId
		case "Description":
			currentStatus = fileSystem.Description
		case "StorageType":
			currentStatus = fileSystem.StorageType
		case "ProtocolType":
			currentStatus = fileSystem.ProtocolType
		case "CreateTime":
			currentStatus = fileSystem.CreateTime
		case "RegionId":
			currentStatus = fileSystem.RegionId
		case "ZoneId":
			currentStatus = fileSystem.ZoneId
		case "FileSystemType":
			currentStatus = fileSystem.FileSystemType
		case "ThroughputMode":
			currentStatus = fileSystem.ThroughputMode
		case "ChargeType":
			currentStatus = fileSystem.ChargeType
		case "KMSKeyId":
			currentStatus = fileSystem.KMSKeyId
		case "ResourceGroupId":
			currentStatus = fileSystem.ResourceGroupId
		case "VpcId":
			currentStatus = fileSystem.VpcId
		case "SnapshotId":
			currentStatus = fileSystem.SnapshotId
		case "FileSystemSpec":
			currentStatus = fileSystem.FileSystemSpec
		case "Version":
			currentStatus = fileSystem.Version
		case "Capacity":
			currentStatus = fmt.Sprintf("%d", fileSystem.Capacity)
		case "UsedCapacity":
			currentStatus = fmt.Sprintf("%d", fileSystem.UsedCapacity)
		case "BandWidth":
			currentStatus = fmt.Sprintf("%d", fileSystem.BandWidth)
		case "EncryptType":
			currentStatus = fmt.Sprintf("%d", fileSystem.EncryptType)
		default:
			// Handle special field patterns
			if strings.HasPrefix(field, "#") {
				// For fields starting with #, check if the field exists and is not empty
				trimmedField := strings.TrimPrefix(field, "#")
				switch trimmedField {
				case "Tags":
					if fileSystem.Tags != nil && len(fileSystem.Tags) > 0 {
						currentStatus = "#CHECKSET"
					}
				case "MountTargets":
					if fileSystem.MountTargets != nil && len(fileSystem.MountTargets) > 0 {
						currentStatus = "#CHECKSET"
					}
				case "AccessGroups":
					if fileSystem.AccessGroups != nil && len(fileSystem.AccessGroups) > 0 {
						currentStatus = "#CHECKSET"
					}
				case "Packages":
					if fileSystem.Packages != nil && len(fileSystem.Packages) > 0 {
						currentStatus = "#CHECKSET"
					}
				case "AutoSnapshotPolicy":
					if fileSystem.AutoSnapshotPolicy != nil {
						currentStatus = "#CHECKSET"
					}
				case "LdapConfig":
					if fileSystem.LdapConfig != nil {
						currentStatus = "#CHECKSET"
					}
				default:
					// For unknown fields, fallback to empty status
					currentStatus = ""
				}
			} else {
				// For unknown field names, return the Status field as default
				currentStatus = fileSystem.Status
			}
		}

		// Check for fail states
		for _, failState := range failStates {
			if currentStatus == failState {
				return fileSystem, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}

		return fileSystem, currentStatus, nil
	}
}

// DescribeAsyncDescribeFileSystems gets async file system information using CWS-Lib-Go API
func (s *NasService) DescribeAsyncDescribeFileSystems(d *schema.ResourceData, res map[string]interface{}) (fileSystem *aliyunNasAPI.FileSystem, err error) {
	return s.DescribeNasFileSystem(d.Id())
}

// ModifyFileSystem modifies a NAS file system using strongly typed parameters
func (s *NasService) ModifyFileSystem(fileSystemId string, description string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := aliyunNasAPI.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.ModifyFileSystem(fileSystemId, description)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ModifyFileSystem", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpgradeFileSystem upgrades the capacity of a NAS file system using direct RPC call
func (s *NasService) UpgradeFileSystem(fileSystemId string, capacity int64) error {
	client := s.client
	action := "UpgradeFileSystem"
	var response map[string]interface{}
	var err error

	request := map[string]interface{}{
		"FileSystemId": fileSystemId,
		"Capacity":     capacity,
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, nil, request, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *NasService) CreateNasFileSystem(fileSystem *aliyunNasAPI.FileSystem) (*aliyunNasAPI.FileSystem, error) {
	nasAPI := s.aliyunNasAPI

	// Create file system using CWS-Lib-Go API
	createdFileSystem, err := nasAPI.CreateFileSystem(fileSystem)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_file_system", "CreateFileSystem", AlibabaCloudSdkGoERROR)
	}

	return createdFileSystem, nil
}

func (s *NasService) DeleteNasFileSystem(fileSystemId string) error {
	nasAPI := s.aliyunNasAPI

	err := nasAPI.DeleteFileSystem(fileSystemId)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "DeleteFileSystem", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListNasFileSystems lists all NAS file systems using CWS-Lib-Go API
func (s *NasService) ListNasFileSystems() ([]*aliyunNasAPI.FileSystem, error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := aliyunNasAPI.NewNasAPI(credentials)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_file_systems", "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	fileSystems, err := nasAPI.ListFileSystems()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_file_systems", "ListFileSystems", AlibabaCloudSdkGoERROR)
	}

	// Convert slice to slice of pointers for consistency
	result := make([]*aliyunNasAPI.FileSystem, len(fileSystems))
	for i := range fileSystems {
		result[i] = &fileSystems[i]
	}

	return result, nil
}
