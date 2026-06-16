package alicloud

import (
	"fmt"
	"testing"
)

func TestIsAlreadyExistError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "alikafka topic already exists",
			err:  fmt.Errorf("Code: BIZ_TOPIC_ALREADY_EXISTS, Message: specified topic already exists.. Please check and try again later"),
		},
		{
			name: "log machine group already exists",
			err:  fmt.Errorf("Code: MachineGroupAlreadyExist, Message: MachineGroup log-rund-pai-sale already exist"),
		},
		{
			name: "ots table already exists",
			err:  fmt.Errorf("OTSObjectAlreadyExist Requested table already exists"),
		},
		{
			name: "alikafka consumer group already exists",
			err:  fmt.Errorf("Code: BIZ_SUBSCRIPTION_ALREADY_EXISTS, Message: specified consumerId already exists. Please check and try again later"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !IsAlreadyExistError(tc.err) {
				t.Fatalf("expected IsAlreadyExistError to return true for error: %v", tc.err)
			}
		})
	}
}

func TestIsProjectTransferAccelerationNotSupportedError(t *testing.T) {
	err := WrapErrorf(
		fmt.Errorf("PutProjectTransferAcceleration SDK error in DisableProjectTransferAcceleration: failed to disable transfer acceleration for project test-project: SDKError:\n   StatusCode: 400\n   Code: NotSupported\n   Message: The operation is not supported in this region."),
		DefaultErrorMsg,
		"test-project",
		"DisableProjectTransferAcceleration",
		AlibabaCloudSdkGoERROR,
	)

	if !isProjectTransferAccelerationNotSupportedError(err) {
		t.Fatalf("expected transfer acceleration NotSupported error to be recognized: %v", err)
	}
}
