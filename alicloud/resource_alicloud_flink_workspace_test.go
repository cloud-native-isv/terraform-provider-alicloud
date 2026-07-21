package alicloud

import (
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type fakeFlinkWorkspacePostCreateService struct {
	waits int
	id    string
	err   error
}

func (f *fakeFlinkWorkspacePostCreateService) WaitForWorkspaceStarting(id string, _ time.Duration) error {
	f.waits++
	f.id = id
	return f.err
}

func TestFlinkWorkspaceCreateDefersEveryPostIDObserverOutsideCreateAction(t *testing.T) {
	for _, test := range []struct {
		name                string
		usesInitialCapacity bool
		waitErr             error
		readErr             error
	}{
		{name: "initial capacity returns after SetId", usesInitialCapacity: true},
		{name: "legacy successful observer is deferred"},
		{name: "legacy readiness failure is deferred", waitErr: errors.New("workspace did not reach STARTING")},
		{name: "legacy authoritative Read failure is deferred", readErr: errors.New("workspace not yet visible by ID")},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})
			service := &fakeFlinkWorkspacePostCreateService{err: test.waitErr}
			reads := 0
			meta := &struct{}{}
			read := func(gotData *schema.ResourceData, gotMeta interface{}) error {
				reads++
				if gotData != data || gotMeta != meta {
					t.Fatalf("Read received data=%p meta=%p, want data=%p meta=%p", gotData, gotMeta, data, meta)
				}
				return test.readErr
			}

			if err := completeFlinkWorkspaceCreate(data, meta, service, "f-real", test.usesInitialCapacity, time.Minute, read); err != nil {
				t.Fatal(err)
			}
			if data.Id() != "f-real" {
				t.Fatalf("ID = %q, want f-real", data.Id())
			}
			if service.waits != 0 || reads != 0 {
				t.Fatalf("post-ID observer ran inside Create action: waits=%d reads=%d, want 0/0", service.waits, reads)
			}
			if service.id != "" {
				t.Fatalf("WaitForWorkspaceStarting unexpectedly observed ID %q", service.id)
			}
		})
	}
}

func TestFlinkWorkspaceLegacyReadinessObserverRunsOnlyDuringProtectedFirstRefresh(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, resource.Schema, config)
	request, options := flinkWorkspaceTestCreateIntent(config, "LEGACY")
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, "LEGACY"); err != nil {
		t.Fatal(err)
	}
	data.SetId("f-paid")
	service := &fakeFlinkWorkspacePostCreateService{err: errors.New("workspace did not reach STARTING")}

	err := waitForFlinkWorkspaceReadinessBeforeFirstRead(data, service, time.Minute)
	if err == nil || data.Id() != "f-paid" || data.Get("identity_visibility_state") != flinkWorkspaceIdentityAwaitingFirstRead {
		t.Fatalf("readiness error=%v id=%q visibility=%#v, want error with protected paid identity", err, data.Id(), data.Get("identity_visibility_state"))
	}
	if service.waits != 1 || service.id != "f-paid" {
		t.Fatalf("readiness observer calls=%d id=%q, want 1/f-paid", service.waits, service.id)
	}

	if err := data.Set("identity_visibility_state", flinkWorkspaceIdentityStable); err != nil {
		t.Fatal(err)
	}
	service.err = errors.New("must not run after first successful Read")
	if err := waitForFlinkWorkspaceReadinessBeforeFirstRead(data, service, time.Minute); err != nil {
		t.Fatal(err)
	}
	if service.waits != 1 {
		t.Fatalf("stable identity reran readiness observer %d times, want 1 total", service.waits)
	}
}
