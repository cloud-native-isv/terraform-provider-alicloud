package cbn

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DeregisterTransitRouterMulticastGroupSources invokes the cbn.DeregisterTransitRouterMulticastGroupSources API synchronously
func (client *Client) DeregisterTransitRouterMulticastGroupSources(request *DeregisterTransitRouterMulticastGroupSourcesRequest) (response *DeregisterTransitRouterMulticastGroupSourcesResponse, err error) {
	response = CreateDeregisterTransitRouterMulticastGroupSourcesResponse()
	err = client.DoAction(request, response)
	return
}

// DeregisterTransitRouterMulticastGroupSourcesWithChan invokes the cbn.DeregisterTransitRouterMulticastGroupSources API asynchronously
func (client *Client) DeregisterTransitRouterMulticastGroupSourcesWithChan(request *DeregisterTransitRouterMulticastGroupSourcesRequest) (<-chan *DeregisterTransitRouterMulticastGroupSourcesResponse, <-chan error) {
	responseChan := make(chan *DeregisterTransitRouterMulticastGroupSourcesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeregisterTransitRouterMulticastGroupSources(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DeregisterTransitRouterMulticastGroupSourcesWithCallback invokes the cbn.DeregisterTransitRouterMulticastGroupSources API asynchronously
func (client *Client) DeregisterTransitRouterMulticastGroupSourcesWithCallback(request *DeregisterTransitRouterMulticastGroupSourcesRequest, callback func(response *DeregisterTransitRouterMulticastGroupSourcesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeregisterTransitRouterMulticastGroupSourcesResponse
		var err error
		defer close(result)
		response, err = client.DeregisterTransitRouterMulticastGroupSources(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DeregisterTransitRouterMulticastGroupSourcesRequest is the request struct for api DeregisterTransitRouterMulticastGroupSources
type DeregisterTransitRouterMulticastGroupSourcesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId                requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ClientToken                    string           `position:"Query" name:"ClientToken"`
	NetworkInterfaceIds            *[]string        `position:"Query" name:"NetworkInterfaceIds"  type:"Repeated"`
	TransitRouterMulticastDomainId string           `position:"Query" name:"TransitRouterMulticastDomainId"`
	ConnectPeerIds                 *[]string        `position:"Query" name:"ConnectPeerIds"  type:"Repeated"`
	GroupIpAddress                 string           `position:"Query" name:"GroupIpAddress"`
	DryRun                         requests.Boolean `position:"Query" name:"DryRun"`
	ResourceOwnerAccount           string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount                   string           `position:"Query" name:"OwnerAccount"`
	OwnerId                        requests.Integer `position:"Query" name:"OwnerId"`
	Version                        string           `position:"Query" name:"Version"`
}

// DeregisterTransitRouterMulticastGroupSourcesResponse is the response struct for api DeregisterTransitRouterMulticastGroupSources
type DeregisterTransitRouterMulticastGroupSourcesResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeregisterTransitRouterMulticastGroupSourcesRequest creates a request to invoke DeregisterTransitRouterMulticastGroupSources API
func CreateDeregisterTransitRouterMulticastGroupSourcesRequest() (request *DeregisterTransitRouterMulticastGroupSourcesRequest) {
	request = &DeregisterTransitRouterMulticastGroupSourcesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Cbn", "2017-09-12", "DeregisterTransitRouterMulticastGroupSources", "", "")
	request.Method = requests.POST
	return
}

// CreateDeregisterTransitRouterMulticastGroupSourcesResponse creates a response to parse from DeregisterTransitRouterMulticastGroupSources response
func CreateDeregisterTransitRouterMulticastGroupSourcesResponse() (response *DeregisterTransitRouterMulticastGroupSourcesResponse) {
	response = &DeregisterTransitRouterMulticastGroupSourcesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}