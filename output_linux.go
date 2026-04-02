//go:build linux

package main

import (
	"os"
	"syscall"
	"time"
)

// getAccessTime returns the access time from FileInfo (Linux version)
func getAccessTime(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(sys.Atim.Sec, sys.Atim.Nsec)
	}
	return info.ModTime()
}

// getCreateTime returns the creation time from FileInfo (Linux version)
func getCreateTime(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(sys.Ctim.Sec, sys.Ctim.Nsec)
	}
	return info.ModTime()
}
