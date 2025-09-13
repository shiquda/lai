//go:build !windows

package display

// enableWindowsANSI is a no-op on non-Windows platforms
func enableWindowsANSI() bool {
	return true
}
