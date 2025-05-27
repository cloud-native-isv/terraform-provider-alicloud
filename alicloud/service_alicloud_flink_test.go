package alicloud

import (
	"os"
	"testing"
)

// TestAccFlinkService_DescribeSupportedZones tests the DescribeSupportedZones function
func TestAccFlinkService_DescribeSupportedZones(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test, set TF_ACC=1 to run")
	}

}
