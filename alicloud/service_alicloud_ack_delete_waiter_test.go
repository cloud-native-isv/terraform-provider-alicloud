package alicloud

import (
	"errors"
	"strings"
	"testing"
	"time"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// These tests pin the delete-waiter semantics of the three cs resources
// (cluster / node pool / addon) offline: stub refresh functions drive the
// real SDK v1.17.2 WaitForState loop (reviewer finding B1).
//
// Contract being pinned:
//   - the delete waiters use an empty Target, the fork's "absence-wait"
//     idiom (see resourceAliCloudFlinkNamespaceDelete), so a nil refresh
//     result means the object is gone and the destroy succeeds;
//   - the delete refresh functions translate a NotFound-classified Describe
//     error into (nil, "", nil); that same classification is the condition
//     the Read functions rely on to SetId("");
//   - Pending stays nil, so any intermediate state keeps polling until the
//     object disappears or the waiter times out.

// ackWaiterObservation is one refresh result while the object still exists.
type ackWaiterObservation struct {
	result interface{}
	state  string
}

// ackDeleteWaiterRefreshStub emulates the fixed delete refresh functions
// (AckClusterDeleteStateRefreshFunc and friends): it replays the
// observations recorded while the object existed, then permanently reports
// the Describe error the ack API layer returns once the object is gone —
// (nil, "", nil) when the error classifies as NotFound, the wrapped error
// otherwise.
func ackDeleteWaiterRefreshStub(observations []ackWaiterObservation, apiErr error) resource.StateRefreshFunc {
	calls := 0
	return func() (interface{}, string, error) {
		if calls < len(observations) {
			obs := observations[calls]
			calls++
			return obs.result, obs.state, nil
		}
		calls++
		if NotFoundError(apiErr) {
			return nil, "", nil
		}
		return nil, "", WrapError(apiErr)
	}
}

// TestAckServiceDeleteWaiterAbsenceIdiom drives WaitForState with the exact
// waiter shape of the three Delete paths (buildAckStateConf, Pending nil,
// Target empty) and asserts the destroy returns nil once the resource is
// absent, no matter whether it vanished immediately or after intermediate
// deleting states. Production delay/poll values are shrunk to milliseconds
// so the tests stay offline and fast.
func TestAckServiceDeleteWaiterAbsenceIdiom(t *testing.T) {
	clusterNotFound := aliyunAckAPI.NewAckServiceError("req-delete-waiter-cluster", "cluster not found")
	for _, tc := range []struct {
		name         string
		apiErr       error
		observations []ackWaiterObservation
		wantErr      string // empty means the destroy must succeed
	}{
		{
			name:   "cluster already absent: DescribeClusterDetail returns NotFound",
			apiErr: clusterNotFound,
		},
		{
			name:   "cluster running then deleting then absent",
			apiErr: clusterNotFound,
			observations: []ackWaiterObservation{
				{result: &aliyunAckAPI.AckCluster{ClusterId: "c-demo", State: "running"}, state: "running"},
				{result: &aliyunAckAPI.AckCluster{ClusterId: "c-demo", State: "deleting"}, state: "deleting"},
			},
		},
		{
			name:   "node pool already absent: DescribeClusterNodePoolDetail returns NotFound",
			apiErr: aliyunAckAPI.NewAckAPIError("DescribeClusterNodePoolDetail", "nodepool not found", nil),
		},
		{
			name:   "node pool deleting then absent",
			apiErr: aliyunAckAPI.NewAckAPIError("DescribeClusterNodePoolDetail", "nodepool not found", nil),
			observations: []ackWaiterObservation{
				{result: &aliyunAckAPI.AckNodePool{ClusterId: "c-demo", NodePoolId: "np-demo", Status: "deleting"}, state: "deleting"},
			},
		},
		{
			name:   "addon already absent: DescribeClusterAddonInstance returns NotFound",
			apiErr: aliyunAckAPI.NewAckSDKError("cs", "DescribeClusterAddonInstance", "addon not exist", nil),
		},
		{
			name:   "addon deleting then absent",
			apiErr: aliyunAckAPI.NewAckSDKError("cs", "DescribeClusterAddonInstance", "addon not exist", nil),
			observations: []ackWaiterObservation{
				{result: &aliyunAckAPI.AckAddonInstance{Name: "terway-eniip", State: "deleting"}, state: "deleting"},
			},
		},
		{
			name:    "non-NotFound Describe error still fails the destroy",
			apiErr:  aliyunAckAPI.NewAckAPIError("DescribeClusterDetail", "internal server error", nil),
			wantErr: "internal server error",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// For the success cases the vanished-object Describe error must
			// classify as NotFound: that classification is exactly what the
			// fixed refresh functions use to return (nil, "", nil) and what
			// the Read functions use to SetId("").
			if tc.wantErr == "" && !NotFoundError(tc.apiErr) {
				t.Fatalf("NotFoundError(%v) = false, want true", tc.apiErr)
			}
			conf := buildAckStateConf(
				nil, []string{},
				10*time.Second, time.Millisecond, time.Millisecond,
				ackDeleteWaiterRefreshStub(tc.observations, tc.apiErr))
			result, err := conf.WaitForState()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("delete waiter error = %v, want nil (destroy must not false-fail)", err)
				}
				if result != nil {
					t.Fatalf("delete waiter result = %v, want nil", result)
				}
				return
			}
			if err == nil {
				t.Fatalf("delete waiter error = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("delete waiter error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// TestAckServiceDeleteWaiterNonEmptyTargetFalseFails pins the SDK v1.17.2
// semantics behind B1: with a non-empty Target a nil refresh result is only
// counted by notfoundTick (the reported state string is ignored), and after
// the default NotFoundChecks=20 the waiter fails with NotFoundError even
// though the object is already gone. This documents why the cs delete
// waiters must keep Target empty.
func TestAckServiceDeleteWaiterNonEmptyTargetFalseFails(t *testing.T) {
	// Pre-fix idiom: NotFound mapped onto a synthetic "deleted" target state.
	refresh := func() (interface{}, string, error) { return nil, "deleted", nil }
	conf := buildAckStateConf(
		nil, []string{"deleted"},
		10*time.Second, time.Millisecond, time.Millisecond, refresh)
	_, err := conf.WaitForState()
	var notFound *resource.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("WaitForState error = %v, want *resource.NotFoundError", err)
	}
	if notFound.Retries != 21 {
		t.Fatalf("NotFoundError.Retries = %d, want 21 (default NotFoundChecks = 20)", notFound.Retries)
	}
}
