package main

import "runtime"

// Colors holds ANSI color codes for terminal output
type Colors struct {
    Red    string
    Green  string
    Yellow string
    Blue   string
    Reset  string
}

// colors is the global color configuration
var colors = getColors()

// getColors returns the color codes based on the OS
// Windows CMD doesn't support ANSI colors by default, but modern Windows Terminal does
func getColors() Colors {

    // Check if running in a terminal that supports colors
    if runtime.GOOS == "windows" {
        // Enable virtual terminal processing on Windows
        enableWindowsColors()
    }

    return Colors{
        Red:    "\033[0;31m",
        Green:  "\033[0;32m",
        Yellow: "\033[1;33m",
        Blue:   "\033[0;34m",
        Reset:  "\033[0m",
    }
}
