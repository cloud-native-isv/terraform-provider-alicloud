package alicloud

import (
	"time"

	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateNasFileset creates a new NAS fileset
func (s *NasService) CreateNasFileset(fileSystemId, fileSystemPath, description string, deletionProtection bool) (*aliyunNasAPI.Fileset, error) {
	nasAPI := s.GetAPI()

	fileset, err := nasAPI.CreateFileset(fileSystemId, fileSystemPath, description, deletionProtection)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_fileset", "CreateFileset", AlibabaCloudSdkGoERROR)
	}

	return fileset, nil
}

// DescribeNasFileset gets NAS fileset information
func (s *NasService) DescribeNasFileset(id string) (*aliyunNasAPI.Fileset, error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	fileSystemId := parts[0]
	fsetId := parts[1]

	nasAPI := s.GetAPI()

	fileset, err := nasAPI.GetFileset(fileSystemId, fsetId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound", "InvalidFileset.NotFound"}) {
			return nil, WrapErrorf(NotFoundErr("NAS:Fileset", id), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetFileset", AlibabaCloudSdkGoERROR)
	}

	return fileset, nil
}

// UpdateNasFileset updates a NAS fileset
func (s *NasService) UpdateNasFileset(fileSystemId, fsetId, description string, deletionProtection bool) error {
	nasAPI := s.GetAPI()

	err := nasAPI.ModifyFileset(fileSystemId, fsetId, description, deletionProtection)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fsetId, "ModifyFileset", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNasFileset deletes a NAS fileset
func (s *NasService) DeleteNasFileset(fileSystemId, fsetId string) error {
	nasAPI := s.GetAPI()

	err := nasAPI.DeleteFileset(fileSystemId, fsetId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound", "InvalidFileset.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, fsetId, "DeleteFileset", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListNasFilesets lists all filesets in a file system
func (s *NasService) ListNasFilesets(fileSystemId string) ([]aliyunNasAPI.Fileset, error) {
	nasAPI := s.GetAPI()

	filesets, err := nasAPI.ListFilesets(fileSystemId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fileSystemId, "ListFilesets", AlibabaCloudSdkGoERROR)
	}

	return filesets, nil
}

// NasFilesetStateRefreshFunc returns a StateRefreshFunc for NAS fileset status
func (s *NasService) NasFilesetStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fileset, err := s.DescribeNasFileset(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fileset.Status == failState {
				return fileset, fileset.Status, WrapError(Error(FailedToReachTargetStatus, fileset.Status))
			}
		}

		return fileset, fileset.Status, nil
	}
}

// WaitForNasFileset waits for NAS fileset to reach target status
func (s *NasService) WaitForNasFileset(id string, targetStatus string, timeout int) error {
	stateConf := BuildStateConf([]string{aliyunNasAPI.FilesetStatusCreating}, []string{targetStatus}, time.Duration(timeout)*time.Second, 5*time.Second, s.NasFilesetStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForNasFilesetDeleted waits for NAS fileset to be deleted
func (s *NasService) WaitForNasFilesetDeleted(id string, timeout int) error {
	stateConf := BuildStateConf([]string{aliyunNasAPI.FilesetStatusDeleting}, []string{}, time.Duration(timeout)*time.Second, 5*time.Second, s.NasFilesetStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
