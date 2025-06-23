package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeNasFileSystem gets NAS file system information using CWS-Lib-Go API
func (s *NasService) DescribeNasFileSystem(id string) (object map[string]interface{}, err error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	fileSystem, err := nasAPI.GetFileSystem(id)
	if err != nil {
		if nas.IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("FileSystem", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetFileSystem", AlibabaCloudSdkGoERROR)
	}

	// Convert FileSystem struct to map[string]interface{} for compatibility
	result := map[string]interface{}{
		"FileSystemId":    fileSystem.FileSystemId,
		"Description":     fileSystem.Description,
		"StorageType":     fileSystem.StorageType,
		"ProtocolType":    fileSystem.ProtocolType,
		"CreateTime":      fileSystem.CreateTime,
		"RegionId":        fileSystem.RegionId,
		"ZoneId":          fileSystem.ZoneId,
		"FileSystemType":  fileSystem.FileSystemType,
		"Status":          fileSystem.Status,
		"Capacity":        fileSystem.Capacity,
		"EncryptType":     fileSystem.EncryptType,
		"KMSKeyId":        fileSystem.KMSKeyId,
		"ResourceGroupId": fileSystem.ResourceGroupId,
		"VpcId":           fileSystem.VpcId,
		"VSwitchId":       fileSystem.VSwitchId,
	}

	return result, nil
}

// NasFileSystemStateRefreshFunc returns a StateRefreshFunc for NAS file system status
func (s *NasService) NasFileSystemStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeNasFileSystem(id)
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

// DescribeNasFileSystemStateRefreshFunc is an alias for backward compatibility
func (s *NasService) DescribeNasFileSystemStateRefreshFunc(id string, defaultRetryState string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeNasFileSystem(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, defaultRetryState, nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if fmt.Sprint(object["Status"]) == failState {
				return object, fmt.Sprint(object["Status"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["Status"])))
			}
		}
		return object, fmt.Sprint(object["Status"]), nil
	}
}

// DescribeAsyncNasFileSystemStateRefreshFunc returns an async StateRefreshFunc for NAS file system
func (s *NasService) DescribeAsyncNasFileSystemStateRefreshFunc(d *schema.ResourceData, res map[string]interface{}, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAsyncDescribeFileSystems(d, res)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
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
				if _err, ok := object["error"]; ok {
					return _err, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
				}
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeAsyncDescribeFileSystems gets async file system information using CWS-Lib-Go API
func (s *NasService) DescribeAsyncDescribeFileSystems(d *schema.ResourceData, res map[string]interface{}) (object map[string]interface{}, err error) {
	return s.DescribeNasFileSystem(d.Id())
}

// ModifyFileSystem modifies a NAS file system using CWS-Lib-Go API
func (s *NasService) ModifyFileSystem(request map[string]interface{}) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, request["FileSystemId"], "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	fileSystemId := fmt.Sprint(request["FileSystemId"])
	description := ""
	if desc, ok := request["Description"]; ok {
		description = fmt.Sprint(desc)
	}

	err = nasAPI.ModifyFileSystem(fileSystemId, description)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ModifyFileSystem", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpgradeFileSystem upgrades the capacity of a NAS file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) UpgradeFileSystem(request map[string]interface{}) error {
	client := s.client
	action := "UpgradeFileSystem"
	var response map[string]interface{}
	var err error

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
		return WrapErrorf(err, DefaultErrorMsg, request["FileSystemId"], action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// EnableRecycleBin enables the recycle bin for a NAS file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) EnableRecycleBin(request map[string]interface{}) error {
	client := s.client
	action := "EnableRecycleBin"
	var response map[string]interface{}
	var err error

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
		return WrapErrorf(err, DefaultErrorMsg, request["FileSystemId"], action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DisableAndCleanRecycleBin disables and cleans the recycle bin for a NAS file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) DisableAndCleanRecycleBin(fileSystemId string) error {
	client := s.client
	action := "DisableAndCleanRecycleBin"
	var response map[string]interface{}
	var err error

	request := map[string]interface{}{
		"FileSystemId": fileSystemId,
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

// UpdateRecycleBinAttribute updates the recycle bin attributes for a NAS file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) UpdateRecycleBinAttribute(request map[string]interface{}) error {
	client := s.client
	action := "UpdateRecycleBinAttribute"
	var response map[string]interface{}
	var err error

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
		return WrapErrorf(err, DefaultErrorMsg, request["FileSystemId"], action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// CreateNasFileSystem creates a NAS file system using CWS-Lib-Go API
func (s *NasService) CreateNasFileSystem(fileSystem *nas.FileSystem) (*nas.FileSystem, error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_file_system", "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	// Create file system using CWS-Lib-Go API
	createdFileSystem, err := nasAPI.CreateFileSystem(fileSystem)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_file_system", "CreateFileSystem", AlibabaCloudSdkGoERROR)
	}

	return createdFileSystem, nil
}
