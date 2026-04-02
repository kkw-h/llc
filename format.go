package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// formatMode converts file mode to ls-style string
func formatMode(mode os.FileMode, path string) string {
	var result strings.Builder

	switch {
	case mode&os.ModeDir != 0:
		result.WriteByte('d')
	case mode&os.ModeSymlink != 0:
		result.WriteByte('l')
	default:
		result.WriteByte('-')
	}

	// Owner
	result.WriteByte(ifThen(mode&0o400 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o200 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o100 != 0, 'x', '-'))

	// Group
	result.WriteByte(ifThen(mode&0o040 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o020 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o010 != 0, 'x', '-'))

	// Other
	result.WriteByte(ifThen(mode&0o004 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o002 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o001 != 0, 'x', '-'))

	return result.String()
}

// formatSize formats file size
func formatSize(size int64, human bool) string {
	if !human {
		return fmt.Sprintf("%8d", size)
	}

	units := []string{"B", "K", "M", "G", "T", "P"}
	value := float64(size)
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%8dB", size)
	}
	return fmt.Sprintf("%7.1f%s", value, units[unitIndex])
}

// formatTime formats time according to style
func formatTime(t time.Time, style string) string {
	switch style {
	case "iso":
		return t.Format(dateFormatISO)
	case "long-iso":
		return t.Format(dateFormatLongISO)
	case "full-iso":
		return t.Format(dateFormatFullISO)
	default:
		if t.Year() == time.Now().Year() {
			return t.Format(dateFormatDefault)
		}
		return t.Format(dateFormatDefaultYear)
	}
}

// ifThen returns a if cond is true, otherwise b
func ifThen(cond bool, a, b byte) byte {
	if cond {
		return a
	}
	return b
}
