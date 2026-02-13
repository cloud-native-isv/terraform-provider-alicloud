package alicloud

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Stubs for missing test helpers in this environment
func RandString(n int) string {
	b := make([]byte, (n+1)/2)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func testAccPreCheck(t *testing.T) {
	t.Log("Skipping PreCheck in simplified environment")
}

var testAccProviderFactories map[string]terraform.ResourceProviderFactory

func init() {
	testAccProviderFactories = map[string]terraform.ResourceProviderFactory{
		"alicloud": func() (terraform.ResourceProvider, error) {
			return Provider(), nil
		},
	}
}

// Global provider instance for tests (mock)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
}
