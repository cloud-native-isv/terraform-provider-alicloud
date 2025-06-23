package alicloud

import (
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// EnableNasRecycleBin enables NAS recycle bin for a file system
func (s *NasService) EnableNasRecycleBin(fileSystemId string, reservedDays int64) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(err)
	}

	err = nasAPI.EnableRecycleBin(fileSystemId, reservedDays)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// DisableAndCleanNasRecycleBin disables and cleans NAS recycle bin for a file system
func (s *NasService) DisableAndCleanNasRecycleBin(fileSystemId string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(err)
	}

	err = nasAPI.DisableAndCleanRecycleBin(fileSystemId)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// GetNasRecycleBinAttribute gets NAS recycle bin attributes for a file system
func (s *NasService) GetNasRecycleBinAttribute(fileSystemId string) (*aliyunNasAPI.RecycleBinAttribute, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	attr, err := nasAPI.GetRecycleBinAttribute(fileSystemId)
	if err != nil {
		return nil, WrapError(err)
	}

	return attr, nil
}

// UpdateNasRecycleBinAttribute updates NAS recycle bin attributes for a file system
func (s *NasService) UpdateNasRecycleBinAttribute(fileSystemId string, reservedDays int64) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(err)
	}

	err = nasAPI.UpdateRecycleBinAttribute(fileSystemId, reservedDays)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// ListNasRecycledDirectoriesAndFiles lists recycled directories and files in NAS recycle bin
func (s *NasService) ListNasRecycledDirectoriesAndFiles(fileSystemId, nextToken, fileId string, maxResults int64) ([]aliyunNasAPI.RecycledDirectoryOrFile, string, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, "", WrapError(err)
	}

	entries, token, err := nasAPI.ListRecycledDirectoriesAndFiles(fileSystemId, nextToken, fileId, maxResults)
	if err != nil {
		return nil, "", WrapError(err)
	}

	return entries, token, nil
}

// CreateNasRecycleBinRestoreJob creates a restore job for recycled files
func (s *NasService) CreateNasRecycleBinRestoreJob(fileSystemId, fileId, targetFileId, clientToken string) (string, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return "", WrapError(err)
	}

	jobId, err := nasAPI.CreateRecycleBinRestoreJob(fileSystemId, fileId, targetFileId, clientToken)
	if err != nil {
		return "", WrapError(err)
	}

	return jobId, nil
}

// CreateNasRecycleBinDeleteJob creates a delete job for recycled files
func (s *NasService) CreateNasRecycleBinDeleteJob(fileSystemId, fileId, clientToken string) (string, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return "", WrapError(err)
	}

	jobId, err := nasAPI.CreateRecycleBinDeleteJob(fileSystemId, fileId, clientToken)
	if err != nil {
		return "", WrapError(err)
	}

	return jobId, nil
}

// ListNasRecycleBinJobs lists NAS recycle bin jobs
func (s *NasService) ListNasRecycleBinJobs(fileSystemId, jobId, status string, pageNumber, pageSize int64) ([]aliyunNasAPI.RecycleBinJob, int64, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, 0, WrapError(err)
	}

	jobs, totalCount, err := nasAPI.ListRecycleBinJobs(fileSystemId, jobId, status, pageNumber, pageSize)
	if err != nil {
		return nil, 0, WrapError(err)
	}

	return jobs, totalCount, nil
}

// GetNasRecycleBinJob gets details of a specific NAS recycle bin job
func (s *NasService) GetNasRecycleBinJob(fileSystemId, jobId string) (*aliyunNasAPI.RecycleBinJob, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	job, err := nasAPI.GetRecycleBinJob(fileSystemId, jobId)
	if err != nil {
		return nil, WrapError(err)
	}

	return job, nil
}

// CancelNasRecycleBinJob cancels a NAS recycle bin job
func (s *NasService) CancelNasRecycleBinJob(jobId string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapError(err)
	}

	err = nasAPI.CancelRecycleBinJob(jobId)
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
		if aliyunNasAPI.IsNotFoundError(err) {
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
		if aliyunNasAPI.IsNotFoundError(err) {
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
