//go:build flink_legacy_namespace_runtime_fixture && windows
// +build flink_legacy_namespace_runtime_fixture,windows

package alicloud

import (
	"fmt"
	"os"
)

func flinkLegacyRuntimeFixtureLockFile(*os.File, bool) error {
	return fmt.Errorf("the Flink legacy runtime fixture requires flock and is unsupported on Windows")
}

func flinkLegacyRuntimeFixtureUnlockFile(*os.File) error {
	return nil
}
