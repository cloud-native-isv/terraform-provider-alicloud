package alicloud

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type fakeFlinkWorkspacePostCreateService struct {
	waits int
	id    string
}

func (f *fakeFlinkWorkspacePostCreateService) WaitForWorkspaceStarting(id string, _ time.Duration) error {
	f.waits++
	f.id = id
	return nil
}

func TestFlinkWorkspaceCreateSeparatesInitialAndLegacyPostCreatePaths(t *testing.T) {
	for _, test := range []struct {
		name                string
		usesInitialCapacity bool
		wantWaits           int
		wantReads           int
	}{
		{name: "initial capacity returns after SetId", usesInitialCapacity: true},
		{name: "legacy waits and reads", wantWaits: 1, wantReads: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})
			service := &fakeFlinkWorkspacePostCreateService{}
			reads := 0
			meta := &struct{}{}
			read := func(gotData *schema.ResourceData, gotMeta interface{}) error {
				reads++
				if gotData != data || gotMeta != meta {
					t.Fatalf("Read received data=%p meta=%p, want data=%p meta=%p", gotData, gotMeta, data, meta)
				}
				return nil
			}

			if err := completeFlinkWorkspaceCreate(data, meta, service, "f-real", test.usesInitialCapacity, time.Minute, read); err != nil {
				t.Fatal(err)
			}
			if data.Id() != "f-real" {
				t.Fatalf("ID = %q, want f-real", data.Id())
			}
			if service.waits != test.wantWaits || reads != test.wantReads {
				t.Fatalf("waits=%d reads=%d, want waits=%d reads=%d", service.waits, reads, test.wantWaits, test.wantReads)
			}
			if service.waits > 0 && service.id != "f-real" {
				t.Fatalf("WaitForWorkspaceStarting ID = %q, want f-real", service.id)
			}
		})
	}
}
