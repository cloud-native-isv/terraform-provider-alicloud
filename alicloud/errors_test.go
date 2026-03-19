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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !IsAlreadyExistError(tc.err) {
				t.Fatalf("expected IsAlreadyExistError to return true for error: %v", tc.err)
			}
		})
	}
}
