//go:build windows
// +build windows

package alicloud

import (
	"fmt"
	"os"
)

func flinkLegacyTofuRuntimeLockFile(*os.File, bool) error {
	return fmt.Errorf("the Flink legacy runtime fixture requires flock and is unsupported on Windows")
}

func flinkLegacyTofuRuntimeUnlockFile(*os.File) error {
	return nil
}
