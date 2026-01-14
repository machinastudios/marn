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
	// Expand environment variables in command
	command = expandEnvVars(command)

	var cmd *exec.Cmd

	if isWindows() {
		// Use PowerShell with functions for Unix commands
		// Environment variables are already expanded by expandEnvVars() above
		aliasScript := `
function cp {
    param([Parameter(ValueFromRemainingArguments)]$items)
    if ($items.Count -lt 2) {
        Write-Error "cp: missing destination"
        return
    }
    $destination = $items[-1]
    $sources = $items[0..($items.Count-2)]
    $isDestDir = Test-Path -Path $destination -PathType Container -ErrorAction SilentlyContinue
    foreach ($source in $sources) {
        if ([string]::IsNullOrWhiteSpace($source)) {
            Write-Error "cp: source path is empty"
            continue
        }
        if ($isDestDir) {
            Copy-Item -Path $source -Destination $destination -Recurse
        } else {
            Copy-Item -Path $source -Destination $destination -Recurse -Force
        }
    }
}
function mv { Move-Item $args }
function rm { Remove-Item $args }
function cat { Get-Content $args }
function ls { Get-ChildItem $args }
function pwd { (Get-Location).Path }

function mkdir {
    if ($args.Count -eq 0) { return }
    New-Item -ItemType Directory -Force -Path $args | Out-Null
}

function touch {
    foreach ($p in $args) {
        if (-not (Test-Path $p)) {
            New-Item -ItemType File -Force -Path $p | Out-Null
        } else {
            (Get-Item $p).LastWriteTime = Get-Date
        }
    }
}

function grep {
    if ($args.Count -eq 0) { return }
    $pattern = $args[0]
    $files = $args[1..($args.Count-1)]
    if ($files.Count -eq 0) {
        Select-String -Pattern $pattern
    } else {
        Select-String -Pattern $pattern -Path $files
    }
}

function which {
    $cmd = Get-Command $args[0] -ErrorAction SilentlyContinue
    if ($cmd) { $cmd.Source }
}

` + command

		cmd = exec.Command("powershell", "-NoProfile", "-Command", aliasScript)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Dir = currentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
