package alicloud

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Regression: plugin-sdk v1 matches Target only for non-nil refresh results, so
// a nil "instance is gone" result never matched Target=["NotFound"] and the
// destroy waiter always died with "couldn't find resource (21 retries)".
// Target must stay empty (nil result = absence = success).
func TestWaitForOtsInstanceDeleting_NilRefreshCompletes(t *testing.T) {
	refresh := func() (interface{}, string, error) {
		return nil, "NotFound", nil
	}
	if err := waitForOtsInstanceDeleting("ots-test", 30*time.Second, refresh); err != nil {
		t.Fatalf("expected nil-res refresh to complete the waiter, got: %v", err)
	}
}

func TestWaitForOtsInstanceDeleting_DeletingThenGone(t *testing.T) {
	calls := 0
	refresh := func() (interface{}, string, error) {
		calls++
		if calls < 3 {
			return map[string]string{"InstanceStatus": "deleting"}, "deleting", nil
		}
		return nil, "NotFound", nil
	}
	if err := waitForOtsInstanceDeleting("ots-test", 30*time.Second, refresh); err != nil {
		t.Fatalf("expected deleting->gone sequence to complete, got: %v", err)
	}
}

func TestWaitForOtsInstanceDeleting_RefreshErrorPropagates(t *testing.T) {
	sentinel := errors.New("sentinel api failure")
	refresh := func() (interface{}, string, error) {
		return nil, "failed", sentinel
	}
	err := waitForOtsInstanceDeleting("ots-test", 30*time.Second, refresh)
	if err == nil || !strings.Contains(err.Error(), "sentinel api failure") {
		t.Fatalf("expected refresh error to propagate, got: %v", err)
	}
}

// Pins the pre-fix failure mode with the SDK's real StateChangeConf: nil
// refresh results never match a non-empty Target. NotFoundChecks is lowered
// to keep the test fast; the production default is 20 ("21 retries").
func TestStateChangeConf_NilResNeverMatchesNonEmptyTarget(t *testing.T) {
	conf := &resource.StateChangeConf{
		Pending:        []string{"deleting"},
		Target:         []string{"NotFound"},
		Refresh:        func() (interface{}, string, error) { return nil, "NotFound", nil },
		Timeout:        5 * time.Minute,
		Delay:          time.Millisecond,
		MinTimeout:     time.Millisecond,
		NotFoundChecks: 3,
	}
	_, err := conf.WaitForState()
	if err == nil || !strings.Contains(err.Error(), "couldn't find resource") {
		t.Fatalf("expected couldn't-find-resource error, got: %v", err)
	}
}
