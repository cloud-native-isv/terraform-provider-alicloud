package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeNasSmbAcl gets NAS SMB ACL information using direct RPC (not fully available in CWS-Lib-Go yet)
func (s *NasService) DescribeNasSmbAcl(id string) (object map[string]interface{}, err error) {
	// For now, keep using direct RPC since SMB ACL functionality is not yet fully implemented in CWS-Lib-Go
	var response map[string]interface{}
	client := s.client
	action := "DescribeSmbAcl"

	request := map[string]interface{}{
		"FileSystemId": id,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, request, nil, true)
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
		if IsExpectedErrors(err, []string{"Forbidden.NasNotFound", "InvalidFileSystemId.NotFound", "Resource.NotFound"}) {
			err = WrapErrorf(NotFoundErr("NAS:SmbAcl", id), NotFoundMsg, ProviderERROR)
			return object, err
		}
		err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		return object, err
	}
	v, err := jsonpath.Get("$.Acl", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Acl", response)
	}
	object = v.(map[string]interface{})
	if fmt.Sprint(object["Enabled"]) == "false" {
		return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
	}
	return object, nil
}

// DescribeFileSystemDescribeSmbAcl gets file system SMB ACL information using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) DescribeFileSystemDescribeSmbAcl(id string) (object map[string]interface{}, err error) {
	// For now, keep using direct RPC since SMB ACL functionality is not yet fully implemented in CWS-Lib-Go
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]interface{}
	request = make(map[string]interface{})
	query = make(map[string]interface{})
	request["FileSystemId"] = id

	action := "DescribeSmbAcl"

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, query, request, true)

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
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Acl", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Acl", response)
	}

	return v.(map[string]interface{}), nil
}

// DescribeFileSystemDescribeNfsAcl gets file system NFS ACL information using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) DescribeFileSystemDescribeNfsAcl(id string) (object map[string]interface{}, err error) {
	// For now, keep using direct RPC since NFS ACL functionality is not yet fully implemented in CWS-Lib-Go
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]interface{}
	request = make(map[string]interface{})
	query = make(map[string]interface{})
	request["FileSystemId"] = id

	action := "DescribeNfsAcl"

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, query, request, true)

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
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Acl", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Acl", response)
	}

	return v.(map[string]interface{}), nil
}

// EnableNfsAcl enables NFS ACL for a file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) EnableNfsAcl(fileSystemId string) error {
	// TODO: When CWS-Lib-Go implements NFS ACL functionality, replace this with:
	// credentials := &common.Credentials{...}
	// nasAPI, err := nas.NewNasAPI(credentials)
	// return nasAPI.EnableNfsAcl(fileSystemId)

	client := s.client
	action := "EnableNfsAcl"

	request := map[string]interface{}{
		"FileSystemId": fileSystemId,
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DisableNfsAcl disables NFS ACL for a file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) DisableNfsAcl(fileSystemId string) error {
	// TODO: When CWS-Lib-Go implements NFS ACL functionality, replace this with:
	// credentials := &common.Credentials{...}
	// nasAPI, err := nas.NewNasAPI(credentials)
	// return nasAPI.DisableNfsAcl(fileSystemId)

	client := s.client
	action := "DisableNfsAcl"

	request := map[string]interface{}{
		"FileSystemId": fileSystemId,
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// EnableSmbAcl enables SMB ACL for a file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) EnableSmbAcl(request map[string]interface{}) error {
	// TODO: When CWS-Lib-Go implements SMB ACL functionality, replace this with:
	// credentials := &common.Credentials{...}
	// nasAPI, err := nas.NewNasAPI(credentials)
	// return nasAPI.EnableSmbAcl(request)

	client := s.client
	action := "EnableSmbAcl"

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		fileSystemId := fmt.Sprint(request["FileSystemId"])
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DisableSmbAcl disables SMB ACL for a file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) DisableSmbAcl(fileSystemId string) error {
	// TODO: When CWS-Lib-Go implements SMB ACL functionality, replace this with:
	// credentials := &common.Credentials{...}
	// nasAPI, err := nas.NewNasAPI(credentials)
	// return nasAPI.DisableSmbAcl(fileSystemId)

	client := s.client
	action := "DisableSmbAcl"

	request := map[string]interface{}{
		"FileSystemId": fileSystemId,
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ModifySmbAcl modifies SMB ACL configuration for a file system using direct RPC (not available in CWS-Lib-Go yet)
func (s *NasService) ModifySmbAcl(request map[string]interface{}) error {
	// TODO: When CWS-Lib-Go implements SMB ACL functionality, replace this with:
	// credentials := &common.Credentials{...}
	// nasAPI, err := nas.NewNasAPI(credentials)
	// return nasAPI.ModifySmbAcl(request)

	client := s.client
	action := "ModifySmbAcl"

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		fileSystemId := fmt.Sprint(request["FileSystemId"])
		return WrapErrorf(err, DefaultErrorMsg, fileSystemId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}
