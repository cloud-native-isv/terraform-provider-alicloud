//go:build !windows
// +build !windows

package alicloud

import (
	"os"
	"syscall"
)

func flinkLegacyTofuRuntimeLockFile(file *os.File, exclusive bool) error {
	operation := syscall.LOCK_SH
	if exclusive {
		operation = syscall.LOCK_EX
	}
	return syscall.Flock(int(file.Fd()), operation)
}

func flinkLegacyTofuRuntimeUnlockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
