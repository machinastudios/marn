package main

import (
    "io"
    "os"
    "os/exec"
    "runtime"
)

// isWindows returns true if running on Windows
func isWindows() bool {
    return runtime.GOOS == "windows"
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()

    _, err = io.Copy(destination, source)
    return err
}

// getMvnCommand returns the correct Maven command for the current OS
func getMvnCommand() string {
    mvnCmd := "mvn"

    if isWindows() {

        // Check if mvn.cmd exists
        if _, err := exec.LookPath("mvn.cmd"); err == nil {
            mvnCmd = "mvn.cmd"
        }
    }

    return mvnCmd
}

// runMvnCommand runs a Maven command
func runMvnCommand(args ...string) error {
    mvnCmd := getMvnCommand()

    cmd := exec.Command(mvnCmd, args...)
    cmd.Dir = currentDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

// runShellCommand runs a shell command
func runShellCommand(command string) error {
    var cmd *exec.Cmd

    if isWindows() {
        cmd = exec.Command("cmd", "/C", command)
    } else {
        cmd = exec.Command("bash", "-c", command)
    }

    cmd.Dir = currentDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    return cmd.Run()
}
