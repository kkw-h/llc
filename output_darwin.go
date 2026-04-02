//go:build darwin

package main

import (
	"os"
	"syscall"
	"time"
)

// getAccessTime returns the access time from FileInfo (macOS version)
func getAccessTime(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(sys.Atimespec.Sec, sys.Atimespec.Nsec)
	}
	return info.ModTime()
}

// getCreateTime returns the creation time from FileInfo (macOS version)
func getCreateTime(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(sys.Ctimespec.Sec, sys.Ctimespec.Nsec)
	}
	return info.ModTime()
}
