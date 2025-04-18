package alicloud

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type CrService struct {
	client *connectivity.AliyunClient
}

type crCreateNamespaceRequestPayload struct {
	Namespace struct {
		Namespace string `json:"Namespace"`
	} `json:"Namespace"`
}

type crUpdateNamespaceRequestPayload struct {
	Namespace struct {
		AutoCreate        bool   `json:"AutoCreate"`
		DefaultVisibility string `json:"DefaultVisibility"`
	} `json:"Namespace"`
}

type crDescribeNamespaceResponse struct {
	RequestId string `json:"requestId"`
	Data      struct {
		Namespace struct {
			Namespace         string `json:"namespace"`
			AuthorizeType     string `json:"authorizeType"`
			DefaultVisibility string `json:"defaultVisibility"`
			AutoCreate        bool   `json:"autoCreate"`
			NamespaceStatus   string `json:"namespaceStatus"`
		} `json:"namespace"`
	} `json:"data"`
}

type crDescribeNamespaceListResponse struct {
	RequestId string `json:"requestId"`
	Data      struct {
		Namespace []struct {
			Namespace       string `json:"namespace"`
			AuthorizeType   string `json:"authorizeType"`
			NamespaceStatus string `json:"namespaceStatus"`
		} `json:"namespaces"`
	} `json:"data"`
}

const (
	RepoTypePublic  = "PUBLIC"
	RepoTypePrivate = "PRIVATE"
)

type crCreateRepoRequestPayload struct {
	Repo struct {
		RepoNamespace string `json:"RepoNamespace"`
		RepoName      string `json:"RepoName"`
		Summary       string `json:"Summary"`
		Detail        string `json:"Detail"`
		RepoType      string `json:"RepoType"`
	} `json:"Repo"`
}

type crUpdateRepoRequestPayload struct {
	Repo struct {
		Summary  string `json:"Summary"`
		Detail   string `json:"Detail"`
		RepoType string `json:"RepoType"`
	} `json:"Repo"`
}

type crDescribeRepoResponse struct {
	RequestId string `json:"requestId"`
	Data      struct {
		Repo struct {
			Summary        string `json:"summary"`
			Detail         string `json:"detail"`
			RepoNamespace  string `json:"repoNamespace"`
			RepoName       string `json:"repoName"`
			RepoType       string `json:"repoType"`
			RepoDomainList struct {
				Public   string `json:"public"`
				Internal string `json:"internal"`
				Vpc      string `json:"vpc"`
			}
		} `json:"repo"`
	} `json:"data"`
}

type crDescribeReposResponse struct {
	RequestId string `json:"requestId"`
	Data      struct {
		Repos    []crRepo `json:"repos"`
		Total    int      `json:"total"`
		PageSize int      `json:"pageSize"`
		Page     int      `json:"page"`
	} `json:"data"`
}

type crRepo struct {
	Summary        string `json:"summary"`
	RepoNamespace  string `json:"repoNamespace"`
	RepoName       string `json:"repoName"`
	RepoType       string `json:"repoType"`
	RegionId       string `json:"regionId"`
	RepoDomainList struct {
		Public   string `json:"public"`
		Internal string `json:"internal"`
		Vpc      string `json:"vpc"`
	} `json:"repoDomainList"`
}

type crDescribeRepoTagsResponse struct {
	RequestId string `json:"requestId"`
	Data      struct {
		Tags     []crTag `json:"tags"`
		Total    int     `json:"total"`
		PageSize int     `json:"pageSize"`
		Page     int     `json:"page"`
	} `json:"data"`
}

type crTag struct {
	ImageId     string `json:"imageId"`
	Digest      string `json:"digest"`
	Tag         string `json:"tag"`
	Status      string `json:"status"`
	ImageUpdate int    `json:"imageUpdate"`
	ImageCreate int    `json:"imageCreate"`
	ImageSize   int    `json:"imageSize"`
}

func (c *CrService) DescribeCrNamespace(id string) (*cr.GetNamespaceResponse, error) {
	response := &cr.GetNamespaceResponse{}
	request := cr.CreateGetNamespaceRequest()
	request.RegionId = c.client.RegionId
	request.Namespace = id

	var err error
	raw, err := c.client.WithCrClient(func(crClient *cr.Client) (interface{}, error) {
		return crClient.GetNamespace(request)
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"NAMESPACE_NOT_EXIST"}) {
			return response, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	addDebug(request.GetActionName(), raw, request.RoaRequest, request)
	response, _ = raw.(*cr.GetNamespaceResponse)

	return response, nil
}

func (c *CrService) WaitForCRNamespace(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		object, err := c.DescribeCrNamespace(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		var response crDescribeNamespaceResponse
		err = json.Unmarshal(object.GetHttpContentBytes(), &response)
		if err != nil {
			return WrapError(err)
		}
		if response.Data.Namespace.Namespace == id && status != Deleted {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, response.Data.Namespace.Namespace, id, ProviderERROR)
		}
	}
}

func (c *CrService) DescribeCrRepo(id string) (*cr.GetRepoResponse, error) {
	response := &cr.GetRepoResponse{}
	sli := strings.Split(id, SLASH_SEPARATED)
	repoNamespace := sli[0]
	repoName := sli[1]

	request := cr.CreateGetRepoRequest()
	request.RegionId = c.client.RegionId
	request.RepoNamespace = repoNamespace
	request.RepoName = repoName

	raw, err := c.client.WithCrClient(func(crClient *cr.Client) (interface{}, error) {
		return crClient.GetRepo(request)
	})
	response, _ = raw.(*cr.GetRepoResponse)
	if err != nil {
		if IsExpectedErrors(err, []string{"REPO_NOT_EXIST"}) {
			return response, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	addDebug(request.GetActionName(), raw, request.RoaRequest, request)
	return response, nil
}

func (c *CrService) WaitForCrRepo(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := c.DescribeCrRepo(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		var response crDescribeRepoResponse
		err = json.Unmarshal(object.GetHttpContentBytes(), &response)
		if err != nil {
			return WrapError(err)
		}
		respId := response.Data.Repo.RepoNamespace + SLASH_SEPARATED + response.Data.Repo.RepoName
		if respId == id && status != Deleted {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, respId, id, ProviderERROR)
		}
	}
}

func (c *CrService) InstanceStatusRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := c.DescribeCrEEInstance(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if resp.InstanceStatus == failState {
				return resp, resp.InstanceStatus, WrapError(Error(FailedToReachTargetStatus, resp.InstanceStatus))
			}
		}
		return resp, resp.InstanceStatus, nil
	}
}

func (s *CrService) DescribeCrEndpointAclPolicy(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetInstanceEndpoint"
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"EndpointType": parts[1],
		"InstanceId":   parts[0],
	}
	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
	v, err := jsonpath.Get("$.AclEntries", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AclEntries", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("CR", id), NotFoundWithResponse, response)
	}
	for _, v := range v.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["Entry"]) == parts[2] {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}
	if !idExist {
		return object, WrapErrorf(NotFoundErr("CR", id), NotFoundWithResponse, response)
	}
	return object, nil
}

func (s *CrService) DescribeCrEndpointAclService(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetInstanceEndpoint"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"EndpointType": parts[1],
		"InstanceId":   parts[0],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *CrService) CrEndpointAclServiceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeCrEndpointAclService(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
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

func (s *CrService) DescribeCrChartNamespace(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetChartNamespace"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"RegionId":      s.client.RegionId,
		"InstanceId":    parts[0],
		"NamespaceName": parts[1],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
		if IsExpectedErrors(err, []string{"CHART_NAMESPACE_NOT_EXIST"}) {
			return object, WrapErrorf(NotFoundErr("CR", id), NotFoundWithResponse, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *CrService) DescribeCrChartRepository(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetChartRepository"
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"RepoName":          parts[2],
		"RepoNamespaceName": parts[1],
		"InstanceId":        parts[0],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
		if IsExpectedErrors(err, []string{"CHART_REPO_NOT_EXIST"}) {
			return object, WrapErrorf(NotFoundErr("CR", id), NotFoundWithResponse, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *CrService) DescribeCrChain(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetChain"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"ChainId":    parts[1],
		"InstanceId": parts[0],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
		if IsExpectedErrors(err, []string{"CHAIN_NOT_EXIST"}) {
			return object, WrapErrorf(NotFoundErr("CR:Chain", id), NotFoundMsg, ProviderERROR, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *CrService) GetChain(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetChain"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"ChainId":    parts[1],
		"InstanceId": parts[0],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *CrService) DescribeCrVpcEndpointLinkedVpc(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	action := "GetInstanceVpcEndpoint"
	client := s.client
	parts, err := ParseResourceId(id, 4)
	if err != nil {
		return nil, WrapError(err)
	}

	request := map[string]interface{}{
		"InstanceId": parts[0],
		"ModuleName": parts[3],
	}

	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("cr", "2018-12-01", action, nil, request, true)
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

	resp, err := jsonpath.Get("$.LinkedVpcs", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.LinkedVpcs", response)
	}

	if v, ok := resp.([]interface{}); !ok || len(v) < 1 {
		return object, WrapErrorf(NotFoundErr("CR:VpcEndpointLinkedVpc", id), NotFoundWithResponse, response)
	}

	for _, v := range resp.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["VpcId"]) == parts[1] && fmt.Sprint(v.(map[string]interface{})["VswitchId"]) == parts[2] {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}

	if !idExist {
		return object, WrapErrorf(NotFoundErr("CR:VpcEndpointLinkedVpc", id), NotFoundWithResponse, response)
	}

	return object, nil
}

func (s *CrService) CrVpcEndpointLinkedVpcStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeCrVpcEndpointLinkedVpc(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
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
