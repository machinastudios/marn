//go:build !windows

package main

// enableWindowsColors is a no-op on Unix systems
func enableWindowsColors() {
    // Colors are enabled by default on Unix
}
