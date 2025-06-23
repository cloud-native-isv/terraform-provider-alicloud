package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeNasFileSystem gets NAS file system information using CWS-Lib-Go API
func (s *NasService) DescribeNasFileSystem(id string) (fileSystem *aliyunNasAPI.FileSystem, err error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	fileSystem, err = nasAPI.GetFileSystem(id)
	if err != nil {
		if nas.IsNotFoundError(err) {
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

		// Convert struct to map for jsonpath compatibility
		objectMap := map[string]interface{}{
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
			"UsedCapacity":    fileSystem.UsedCapacity,
			"BandWidth":       fileSystem.BandWidth,
			"ThroughputMode":  fileSystem.ThroughputMode,
			"ChargeType":      fileSystem.ChargeType,
			"EncryptType":     fileSystem.EncryptType,
			"KMSKeyId":        fileSystem.KMSKeyId,
			"ResourceGroupId": fileSystem.ResourceGroupId,
			"VpcId":           fileSystem.VpcId,
			"VSwitchId":       fileSystem.VSwitchId,
			"SnapshotId":      fileSystem.SnapshotId,
			"FileSystemSpec":  fileSystem.FileSystemSpec,
			"Version":         fileSystem.Version,
		}

		// Handle Tags
		if fileSystem.Tags != nil {
			tags := make([]map[string]interface{}, len(fileSystem.Tags))
			for i, tag := range fileSystem.Tags {
				tags[i] = map[string]interface{}{
					"key":   tag.Key,
					"value": tag.Value,
				}
			}
			objectMap["Tags"] = tags
		}

		// Handle MountTargets
		if fileSystem.MountTargets != nil {
			mountTargets := make([]map[string]interface{}, len(fileSystem.MountTargets))
			for i, mt := range fileSystem.MountTargets {
				mountTargets[i] = map[string]interface{}{
					"MountTargetDomain": mt.MountTargetDomain,
					"NetworkType":       mt.NetworkType,
					"AccessGroupName":   mt.AccessGroupName,
					"Status":            mt.Status,
					"VpcId":             mt.VpcId,
					"VSwitchId":         mt.VSwitchId,
				}
			}
			objectMap["MountTargets"] = mountTargets
		}

		// Handle AccessGroups
		if fileSystem.AccessGroups != nil {
			accessGroups := make([]map[string]interface{}, len(fileSystem.AccessGroups))
			for i, ag := range fileSystem.AccessGroups {
				accessGroups[i] = map[string]interface{}{
					"AccessGroupName":  ag.AccessGroupName,
					"AccessGroupType":  ag.AccessGroupType,
					"Description":      ag.Description,
					"CreateTime":       ag.CreateTime,
					"ModifyTime":       ag.ModifyTime,
					"RuleCount":        ag.RuleCount,
					"MountTargetCount": ag.MountTargetCount,
					"FileSystemType":   ag.FileSystemType,
				}
			}
			objectMap["AccessGroups"] = accessGroups
		}

		// Handle Packages
		if fileSystem.Packages != nil {
			packages := make([]map[string]interface{}, len(fileSystem.Packages))
			for i, pkg := range fileSystem.Packages {
				packages[i] = map[string]interface{}{
					"PackageId":   pkg.PackageId,
					"PackageType": pkg.PackageType,
					"Size":        pkg.Size,
					"StartTime":   pkg.StartTime,
					"ExpiredTime": pkg.ExpiredTime,
					"Status":      pkg.Status,
				}
			}
			objectMap["Packages"] = packages
		}

		// Handle AutoSnapshotPolicy
		if fileSystem.AutoSnapshotPolicy != nil {
			objectMap["AutoSnapshotPolicy"] = map[string]interface{}{
				"AutoSnapshotPolicyId": fileSystem.AutoSnapshotPolicy.AutoSnapshotPolicyId,
				"PolicyName":           fileSystem.AutoSnapshotPolicy.PolicyName,
				"CreateTime":           fileSystem.AutoSnapshotPolicy.CreateTime,
				"RepeatWeekdays":       fileSystem.AutoSnapshotPolicy.RepeatWeekdays,
				"TimePoints":           fileSystem.AutoSnapshotPolicy.TimePoints,
				"RetentionDays":        fileSystem.AutoSnapshotPolicy.RetentionDays,
				"Status":               fileSystem.AutoSnapshotPolicy.Status,
				"FileSystemNums":       fileSystem.AutoSnapshotPolicy.FileSystemNums,
				"RegionId":             fileSystem.AutoSnapshotPolicy.RegionId,
			}
		}

		// Handle LdapConfig
		if fileSystem.LdapConfig != nil {
			objectMap["LdapConfig"] = map[string]interface{}{
				"URI":        fileSystem.LdapConfig.URI,
				"BindDN":     fileSystem.LdapConfig.BindDN,
				"SearchBase": fileSystem.LdapConfig.SearchBase,
			}
		}

		v, err := jsonpath.Get(field, objectMap)
		currentStatus := fmt.Sprint(v)

		if strings.HasPrefix(field, "#") {
			v, _ := jsonpath.Get(strings.TrimPrefix(field, "#"), objectMap)
			if v != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return fileSystem, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return fileSystem, currentStatus, nil
	}
}

// DescribeNasFileSystemStateRefreshFunc is an alias for backward compatibility
func (s *NasService) DescribeNasFileSystemStateRefreshFunc(id string, defaultRetryState string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fileSystem, err := s.DescribeNasFileSystem(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, defaultRetryState, nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if fileSystem.Status == failState {
				return fileSystem, fileSystem.Status, WrapError(Error(FailedToReachTargetStatus, fileSystem.Status))
			}
		}
		return fileSystem, fileSystem.Status, nil
	}
}

// DescribeAsyncNasFileSystemStateRefreshFunc returns an async StateRefreshFunc for NAS file system
func (s *NasService) DescribeAsyncNasFileSystemStateRefreshFunc(d *schema.ResourceData, res map[string]interface{}, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fileSystem, err := s.DescribeAsyncDescribeFileSystems(d, res)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Convert struct to map for jsonpath compatibility
		objectMap := map[string]interface{}{
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
			"UsedCapacity":    fileSystem.UsedCapacity,
			"BandWidth":       fileSystem.BandWidth,
			"ThroughputMode":  fileSystem.ThroughputMode,
			"ChargeType":      fileSystem.ChargeType,
			"EncryptType":     fileSystem.EncryptType,
			"KMSKeyId":        fileSystem.KMSKeyId,
			"ResourceGroupId": fileSystem.ResourceGroupId,
			"VpcId":           fileSystem.VpcId,
			"VSwitchId":       fileSystem.VSwitchId,
			"SnapshotId":      fileSystem.SnapshotId,
			"FileSystemSpec":  fileSystem.FileSystemSpec,
			"Version":         fileSystem.Version,
		}

		v, err := jsonpath.Get(field, objectMap)
		currentStatus := fmt.Sprint(v)

		if strings.HasPrefix(field, "#") {
			v, _ := jsonpath.Get(strings.TrimPrefix(field, "#"), objectMap)
			if v != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				if _err, ok := res["error"]; ok {
					return _err, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
				}
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

	nasAPI, err := nas.NewNasAPI(credentials)
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

// EnableRecycleBin enables the recycle bin for a NAS file system using CWS-Lib-Go API
func (s *NasService) EnableRecycleBin(fileSystemId string, reservedDays int64) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.EnableRecycleBin(fileSystemId, reservedDays)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "EnableRecycleBin", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DisableAndCleanRecycleBin disables and cleans the recycle bin for a NAS file system using CWS-Lib-Go API
func (s *NasService) DisableAndCleanRecycleBin(fileSystemId string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DisableAndCleanRecycleBin(fileSystemId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "DisableAndCleanRecycleBin", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateRecycleBinAttribute updates the recycle bin attributes for a NAS file system using CWS-Lib-Go API
func (s *NasService) UpdateRecycleBinAttribute(fileSystemId string, reservedDays int64, status string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.UpdateRecycleBinAttribute(fileSystemId, reservedDays)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "UpdateRecycleBinAttribute", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// CreateNasFileSystem creates a NAS file system using CWS-Lib-Go API
func (s *NasService) CreateNasFileSystem(fileSystem *aliyunNasAPI.FileSystem) (*aliyunNasAPI.FileSystem, error) {
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

// DeleteNasFileSystem deletes a NAS file system using CWS-Lib-Go API
func (s *NasService) DeleteNasFileSystem(fileSystemId string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DeleteFileSystem(fileSystemId)
	if err != nil {
		if nas.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "DeleteFileSystem", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DescribeNasMountTargets gets NAS mount targets information using CWS-Lib-Go API
func (s *NasService) DescribeNasMountTargets(fileSystemId string) ([]*aliyunNasAPI.MountTarget, error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	mountTargets, err := nasAPI.ListMountTargets(fileSystemId)
	if err != nil {
		if nas.IsNotFoundError(err) {
			return []*aliyunNasAPI.MountTarget{}, nil
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ListMountTargets", AlibabaCloudSdkGoERROR)
	}

	// Convert slice to slice of pointers for consistency
	result := make([]*aliyunNasAPI.MountTarget, len(mountTargets))
	for i := range mountTargets {
		result[i] = &mountTargets[i]
	}

	return result, nil
}

// CreateNasMountTarget creates a NAS mount target using CWS-Lib-Go API
func (s *NasService) CreateNasMountTarget(fileSystemId string, mountTarget *aliyunNasAPI.MountTarget) (*aliyunNasAPI.MountTarget, error) {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	createdMountTarget, err := nasAPI.CreateMountTarget(fileSystemId, mountTarget)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "CreateMountTarget", AlibabaCloudSdkGoERROR)
	}

	return createdMountTarget, nil
}

// ModifyNasMountTarget modifies a NAS mount target using CWS-Lib-Go API
func (s *NasService) ModifyNasMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.ModifyMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ModifyMountTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNasMountTarget deletes a NAS mount target using CWS-Lib-Go API
func (s *NasService) DeleteNasMountTarget(fileSystemId, mountTargetDomain string) error {
	// Create NAS API client using CWS-Lib-Go
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DeleteMountTarget(fileSystemId, mountTargetDomain)
	if err != nil {
		if nas.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, "DeleteMountTarget", AlibabaCloudSdkGoERROR)
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

	nasAPI, err := nas.NewNasAPI(credentials)
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
