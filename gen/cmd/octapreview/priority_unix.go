//go:build darwin || linux

package main

import "syscall"

// Rendering should leave interactive work responsive; lowering priority is best effort.
func lowerPriority() {
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 19)
}
