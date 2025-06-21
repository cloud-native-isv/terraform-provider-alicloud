package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeNasRecycleBin gets NAS recycle bin information
func (s *NasService) DescribeNasRecycleBin(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetRecycleBinAttribute"
	request := map[string]interface{}{
		"FileSystemId": id,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcGet("NAS", "2017-06-26", action, request, nil)
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
		if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound"}) {
			return object, WrapErrorf(NotFoundErr("NAS:RecycleBin", id), NotFoundMsg, ProviderERROR, fmt.Sprint(response["RequestId"]))
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.RecycleBinAttribute", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.RecycleBinAttribute", response)
	}
	object = v.(map[string]interface{})
	if fmt.Sprint(object["Status"]) == "Disable" {
		return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
	}

	return object, nil
}

// DescribeFileSystemGetRecycleBinAttribute gets file system recycle bin attribute (v2 version)
func (s *NasService) DescribeFileSystemGetRecycleBinAttribute(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]interface{}
	request = make(map[string]interface{})
	query = make(map[string]interface{})
	request["FileSystemId"] = id

	action := "GetRecycleBinAttribute"

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

	v, err := jsonpath.Get("$.RecycleBinAttribute", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.RecycleBinAttribute", response)
	}

	return v.(map[string]interface{}), nil
}
