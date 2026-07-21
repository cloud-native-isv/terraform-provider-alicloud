//go:build flink_legacy_namespace_runtime_fixture && !windows
// +build flink_legacy_namespace_runtime_fixture,!windows

package alicloud

import (
	"os"
	"syscall"
)

func flinkLegacyRuntimeFixtureLockFile(file *os.File, exclusive bool) error {
	operation := syscall.LOCK_SH
	if exclusive {
		operation = syscall.LOCK_EX
	}
	return syscall.Flock(int(file.Fd()), operation)
}

func flinkLegacyRuntimeFixtureUnlockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
