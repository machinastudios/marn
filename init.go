package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// initMarn installs marn globally
func initMarn() {
    execPath, err := os.Executable()
    if err != nil {
        fmt.Printf("%sError: Could not get executable path%s\n", colors.Red, colors.Reset)
        os.Exit(1)
    }

    execPath, _ = filepath.Abs(execPath)
    execDir := filepath.Dir(execPath)

    fmt.Printf("%sInstalling marn globally...%s\n", colors.Blue, colors.Reset)
    fmt.Println()

    if isWindows() {
        // On Windows, just add current directory to PATH
        initWindows(execDir, execPath)
    } else {
        // On Linux/macOS, copy to /usr/local/bin
        initUnix(execPath)
    }
}

// initWindows adds the executable directory to PATH
func initWindows(execDir, execPath string) {
    fmt.Printf("%sExecutable: %s%s\n", colors.Yellow, execPath, colors.Reset)
    fmt.Println()

    // Add to PATH
    addToWindowsPath(execDir)

    fmt.Println()
    fmt.Printf("%s✓ Installation complete!%s\n", colors.Green, colors.Reset)
    fmt.Println()
    fmt.Println("You can now use 'marn' from anywhere in your system!")
    fmt.Println()
    fmt.Println("Test it with: marn --help")
    fmt.Println()
    fmt.Printf("%sNote: You may need to restart your terminal for PATH changes to take effect.%s\n", colors.Yellow, colors.Reset)
}

// initUnix copies the executable to /usr/local/bin
func initUnix(execPath string) {
    targetDir := "/usr/local/bin"
    target := filepath.Join(targetDir, "marn")

    fmt.Printf("%sSource: %s%s\n", colors.Yellow, execPath, colors.Reset)
    fmt.Printf("%sTarget: %s%s\n", colors.Yellow, target, colors.Reset)
    fmt.Println()

    // Create target directory if it doesn't exist
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        fmt.Printf("%sError: Could not create directory %s%s\n", colors.Red, targetDir, colors.Reset)
        os.Exit(1)
    }

    // Copy the executable
    if err := copyFile(execPath, target); err != nil {

        // Try with sudo
        fmt.Printf("%sInstalling to %s requires sudo permissions%s\n", colors.Yellow, targetDir, colors.Reset)

        cmd := exec.Command("sudo", "cp", execPath, target)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Stdin = os.Stdin

        if err := cmd.Run(); err != nil {
            fmt.Printf("%sInstallation failed%s\n", colors.Red, colors.Reset)
            os.Exit(1)
        }

        // Make it executable
        exec.Command("sudo", "chmod", "+x", target).Run()
    } else {
        os.Chmod(target, 0755)
    }

    fmt.Printf("%s✓ Marn installed to %s%s\n", colors.Green, target, colors.Reset)
    fmt.Println()
    fmt.Printf("%s✓ Installation complete!%s\n", colors.Green, colors.Reset)
    fmt.Println()
    fmt.Println("You can now use 'marn' from anywhere in your system!")
    fmt.Println()
    fmt.Println("Test it with: marn --help")
}

// addToWindowsPath adds a directory to the user's PATH on Windows
func addToWindowsPath(dir string) {

    // Get current user PATH
    cmd := exec.Command("powershell", "-Command",
        "[Environment]::GetEnvironmentVariable('PATH', 'User')")
    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("%sWarning: Could not read current PATH%s\n", colors.Yellow, colors.Reset)
        showManualPathInstructions(dir)
        return
    }

    currentPath := strings.TrimSpace(string(output))

    // Check if directory is already in PATH
    pathDirs := strings.Split(currentPath, ";")
    for _, p := range pathDirs {

        if strings.EqualFold(strings.TrimSpace(p), dir) {
            fmt.Printf("%s✓ Directory already in PATH%s\n", colors.Green, colors.Reset)
            return
        }
    }

    // Add to PATH
    fmt.Printf("%sAdding to PATH: %s%s\n", colors.Blue, dir, colors.Reset)

    var newPath string
    if currentPath == "" {
        newPath = dir
    } else {
        newPath = currentPath + ";" + dir
    }

    // Use PowerShell to set the environment variable
    psCmd := fmt.Sprintf("[Environment]::SetEnvironmentVariable('PATH', '%s', 'User')", newPath)
    cmd = exec.Command("powershell", "-Command", psCmd)

    if err := cmd.Run(); err != nil {
        fmt.Printf("%sWarning: Could not add to PATH automatically: %v%s\n", colors.Yellow, err, colors.Reset)
        showManualPathInstructions(dir)
        return
    }

    fmt.Printf("%s✓ Added to PATH%s\n", colors.Green, colors.Reset)
}

// showManualPathInstructions shows manual instructions for adding to PATH
func showManualPathInstructions(dir string) {
    fmt.Println()
    fmt.Println("Please add manually with this PowerShell command:")
    fmt.Printf("  [Environment]::SetEnvironmentVariable(\"PATH\", $env:PATH + \";%s\", \"User\")\n", dir)
}
