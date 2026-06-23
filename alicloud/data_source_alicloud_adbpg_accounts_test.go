package alicloud

import (
	"testing"
)

func TestAdbpgAccountsDataSourceSchema(t *testing.T) {
	r := dataSourceAliCloudAdbpgAccounts()
	if r == nil {
		t.Fatal("expected non-nil data source")
	}

	s := r.Schema["db_instance_id"]
	if !s.Required {
		t.Fatal("expected db_instance_id to be Required")
	}

	if _, ok := r.Schema["name_regex"]; !ok {
		t.Fatal("expected name_regex filter field")
	}

	accounts, ok := r.Schema["accounts"]
	if !ok {
		t.Fatal("expected accounts output field")
	}
	if !accounts.Computed {
		t.Fatal("expected accounts to be Computed")
	}
}
