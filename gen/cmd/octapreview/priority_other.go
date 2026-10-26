//go:build !darwin && !linux

package main

// The worker and memory limits still apply on platforms without Unix niceness.
func lowerPriority() {}
