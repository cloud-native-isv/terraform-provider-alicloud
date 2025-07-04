package alicloud

import (
	common "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *NasService) EnableNasRecycleBin(fileSystemId string, reservedDays int64) error {
	nasAPI := s.aliyunNasAPI

	err := nasAPI.EnableRecycleBin(fileSystemId, reservedDays)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *NasService) DisableAndCleanNasRecycleBin(fileSystemId string) error {
	nasAPI := s.aliyunNasAPI

	err := nasAPI.DisableAndCleanRecycleBin(fileSystemId)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *NasService) GetNasRecycleBinAttribute(fileSystemId string) (*aliyunNasAPI.RecycleBinAttribute, error) {
	nasAPI := s.aliyunNasAPI

	attr, err := nasAPI.GetRecycleBinAttribute(fileSystemId)
	if err != nil {
		return nil, WrapError(err)
	}

	return attr, nil
}

func (s *NasService) UpdateNasRecycleBinAttribute(fileSystemId string, reservedDays int64) error {
	nasAPI := s.aliyunNasAPI

	err := nasAPI.UpdateRecycleBinAttribute(fileSystemId, reservedDays)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *NasService) ListNasRecycledDirectoriesAndFiles(fileSystemId, nextToken, fileId string, maxResults int64) ([]aliyunNasAPI.RecycledDirectoryOrFile, string, error) {
	nasAPI := s.aliyunNasAPI

	entries, token, err := nasAPI.ListRecycledDirectoriesAndFiles(fileSystemId, nextToken, fileId, maxResults)
	if err != nil {
		return nil, "", WrapError(err)
	}

	return entries, token, nil
}

func (s *NasService) CreateNasRecycleBinRestoreJob(fileSystemId, fileId, targetFileId, clientToken string) (string, error) {
	nasAPI := s.aliyunNasAPI

	jobId, err := nasAPI.CreateRecycleBinRestoreJob(fileSystemId, fileId, targetFileId, clientToken)
	if err != nil {
		return "", WrapError(err)
	}

	return jobId, nil
}

func (s *NasService) CreateNasRecycleBinDeleteJob(fileSystemId, fileId, clientToken string) (string, error) {
	nasAPI := s.aliyunNasAPI

	jobId, err := nasAPI.CreateRecycleBinDeleteJob(fileSystemId, fileId, clientToken)
	if err != nil {
		return "", WrapError(err)
	}

	return jobId, nil
}

func (s *NasService) ListNasRecycleBinJobs(fileSystemId, jobId, status string, pageNumber, pageSize int64) ([]aliyunNasAPI.RecycleBinJob, int64, error) {
	nasAPI := s.aliyunNasAPI

	jobs, totalCount, err := nasAPI.ListRecycleBinJobs(fileSystemId, jobId, status, pageNumber, pageSize)
	if err != nil {
		return nil, 0, WrapError(err)
	}

	return jobs, totalCount, nil
}

func (s *NasService) GetNasRecycleBinJob(fileSystemId, jobId string) (*aliyunNasAPI.RecycleBinJob, error) {
	nasAPI := s.aliyunNasAPI

	job, err := nasAPI.GetRecycleBinJob(fileSystemId, jobId)
	if err != nil {
		return nil, WrapError(err)
	}

	return job, nil
}

func (s *NasService) CancelNasRecycleBinJob(jobId string) error {
	nasAPI := s.aliyunNasAPI

	err := nasAPI.CancelRecycleBinJob(jobId)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// NasRecycleBinStateRefreshFunc returns a StateRefreshFunc for waiting for NAS recycle bin state changes
func (s *NasService) NasRecycleBinStateRefreshFunc(fileSystemId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.GetNasRecycleBinAttribute(fileSystemId)
		if err != nil {
			if NotFoundError(err) {
				// Recycle bin is disabled
				return nil, "Disabled", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Status == failState {
				return object, object.Status, WrapError(Error(FailedToReachTargetStatus, object.Status))
			}
		}

		return object, object.Status, nil
	}
}

// NasRecycleBinJobStateRefreshFunc returns a StateRefreshFunc for waiting for NAS recycle bin job state changes
func (s *NasService) NasRecycleBinJobStateRefreshFunc(fileSystemId, jobId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.GetNasRecycleBinJob(fileSystemId, jobId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Status == failState {
				return object, object.Status, WrapError(Error(FailedToReachTargetStatus, object.Status))
			}
		}

		return object, object.Status, nil
	}
}

// DescribeNasRecycleBin gets NAS recycle bin information
func (s *NasService) DescribeNasRecycleBin(id string) (object map[string]interface{}, err error) {
	attr, err := s.GetNasRecycleBinAttribute(id)
	if err != nil {
		if common.IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("NAS:RecycleBin", id), NotFoundMsg, ProviderERROR, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetRecycleBinAttribute", AlibabaCloudSdkGoERROR)
	}

	object = map[string]interface{}{
		"Status":        attr.Status,
		"ReservedDays":  attr.ReservedDays,
		"Size":          attr.Size,
		"SecondarySize": attr.SecondarySize,
		"ArchiveSize":   attr.ArchiveSize,
		"EnableTime":    attr.EnableTime,
	}

	if attr.Status == "Stop" {
		return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, object)
	}

	return object, nil
}

// DescribeFileSystemGetRecycleBinAttribute gets file system recycle bin attribute (v2 version)
func (s *NasService) DescribeFileSystemGetRecycleBinAttribute(id string) (object map[string]interface{}, err error) {
	attr, err := s.GetNasRecycleBinAttribute(id)
	if err != nil {
		if common.IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("NAS:RecycleBin", id), NotFoundMsg, ProviderERROR, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetRecycleBinAttribute", AlibabaCloudSdkGoERROR)
	}

	object = map[string]interface{}{
		"Status":        attr.Status,
		"ReservedDays":  attr.ReservedDays,
		"Size":          attr.Size,
		"SecondarySize": attr.SecondarySize,
		"ArchiveSize":   attr.ArchiveSize,
		"EnableTime":    attr.EnableTime,
	}

	return object, nil
}
