package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

// installDependencies runs mvn dependency:resolve
func installDependencies() {
    fmt.Printf("%sInstalling dependencies...%s\n", colors.Green, colors.Reset)
    runMvnCommand("dependency:resolve")
}

// linkProject links the project to local Maven repository
func linkProject() {

    if _, err := os.Stat(pomFile); os.IsNotExist(err) {
        fmt.Printf("%sError: pom.xml not found%s\n", colors.Red, colors.Reset)
        fmt.Println("Please run 'marn link' from a Maven project directory.")
        os.Exit(1)
    }

    fmt.Printf("%sLinking project to local Maven repository...%s\n", colors.Blue, colors.Reset)
    fmt.Println()

    fmt.Printf("%sInstalling project to ~/.m2/repository...%s\n", colors.Green, colors.Reset)

    if err := runMvnCommand("clean", "install", "-DskipTests"); err != nil {
        fmt.Printf("%s✗ Failed to link project%s\n", colors.Red, colors.Reset)
        os.Exit(1)
    }

    fmt.Println()
    fmt.Printf("%s✓ Project linked to local Maven repository!%s\n", colors.Green, colors.Reset)
    fmt.Println()
    fmt.Println("Other projects can now use this as a dependency.")
}

// buildProject builds the project
func buildProject() {
    fmt.Printf("%sBuilding project...%s\n", colors.Green, colors.Reset)

    if err := buildLocalDependencies(true); err != nil {
        os.Exit(1)
    }

    runMvnCommand("clean", "compile")
}

// testProject runs tests
func testProject() {
    fmt.Printf("%sRunning tests...%s\n", colors.Green, colors.Reset)

    if err := buildLocalDependencies(true); err != nil {
        os.Exit(1)
    }

    runMvnCommand("test")
}

// packageProject packages the project
func packageProject() {
    fmt.Printf("%sPackaging project...%s\n", colors.Green, colors.Reset)
    runMvnCommand("clean", "package")
}

// cleanProject cleans the project
func cleanProject() {
    fmt.Printf("%sCleaning project...%s\n", colors.Green, colors.Reset)
    runMvnCommand("clean")
}

// runProject builds and runs the JAR
func runProject() {
    fmt.Printf("%sBuilding and running project...%s\n", colors.Green, colors.Reset)

    if err := buildLocalDependencies(true); err != nil {
        os.Exit(1)
    }

    // Get artifact ID and main class
    artifactID := getArtifactID()
    mainClass := getMainClass()

    // Kill existing processes
    killExistingProcesses(artifactID, mainClass)

    // Create data directory if needed
    os.MkdirAll("data", 0755)

    // Build the project
    if err := runMvnCommand("clean", "package", "-DskipTests"); err != nil {
        os.Exit(1)
    }

    // Find the JAR file
    jarFile := findJarFile()
    if jarFile == "" {
        fmt.Printf("%sError: No JAR file found in target/ directory%s\n", colors.Red, colors.Reset)
        os.Exit(1)
    }

    fmt.Printf("%sRunning: %s%s\n", colors.Green, jarFile, colors.Reset)
    fmt.Println()

    // Run the JAR with additional arguments
    args := []string{"-jar", jarFile}
    if len(os.Args) > 2 {
        args = append(args, os.Args[2:]...)
    }

    cmd := exec.Command("java", args...)
    cmd.Dir = currentDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    cmd.Run()
}

// executeScript executes a custom script from pom.xml
func executeScript(scriptName string) {
    scripts := getScriptsFromPom()

    content, exists := scripts[scriptName]
    if !exists {
        fmt.Printf("%sError: Script '%s' not found in pom.xml%s\n", colors.Red, scriptName, colors.Reset)
        fmt.Println()
        listScripts()
        os.Exit(1)
    }

    // Expand environment variables in script content
    content = expandEnvVars(content)

    fmt.Printf("%sExecuting script: %s%s\n", colors.Yellow, scriptName, colors.Reset)
    fmt.Printf("%sCommand: %s%s\n", colors.Blue, content, colors.Reset)
    fmt.Println()

    runShellCommand(content)
}

// buildLocalDependencies builds all local dependencies
func buildLocalDependencies(skipTests bool) error {
    deps := getLocalDependencies()

    if len(deps) == 0 {
        return nil
    }

    fmt.Printf("%sBuilding local dependencies first...%s\n", colors.Yellow, colors.Reset)

    for _, depPath := range deps {
        pomPath := filepath.Join(depPath, "pom.xml")

        if _, err := os.Stat(pomPath); err == nil {
            fmt.Printf("%sBuilding dependency: %s%s\n", colors.Blue, depPath, colors.Reset)

            args := []string{"clean", "install"}
            if skipTests {
                args = append(args, "-DskipTests")
            }

            cmd := exec.Command("mvn", args...)
            cmd.Dir = depPath
            cmd.Stdout = io.Discard
            cmd.Stderr = io.Discard

            if err := cmd.Run(); err != nil {
                fmt.Printf("%sFailed to build dependency: %s%s\n", colors.Red, depPath, colors.Reset)
                return err
            }

            fmt.Printf("%s✓ Dependency built: %s%s\n", colors.Green, depPath, colors.Reset)
        }
    }

    fmt.Println()
    return nil
}

// findJarFile finds the first JAR file in target directory
func findJarFile() string {
    targetDir := filepath.Join(currentDir, "target")

    entries, err := os.ReadDir(targetDir)
    if err != nil {
        return ""
    }

    for _, entry := range entries {

        if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jar") {
            return filepath.Join(targetDir, entry.Name())
        }
    }

    return ""
}

// killExistingProcesses kills existing Java processes
func killExistingProcesses(artifactID, mainClass string) {

    if isWindows() {

        // On Windows, use taskkill
        if mainClass != "" {
            exec.Command("taskkill", "/F", "/FI", fmt.Sprintf("WINDOWTITLE eq *%s*", mainClass)).Run()
        }

        if artifactID != "" {
            exec.Command("taskkill", "/F", "/FI", fmt.Sprintf("WINDOWTITLE eq *%s*", artifactID)).Run()
        }
    } else {

        // On Unix, use pkill
        if mainClass != "" {
            exec.Command("pkill", "-f", fmt.Sprintf("java.*%s", mainClass)).Run()
        }

        if artifactID != "" {
            exec.Command("pkill", "-f", fmt.Sprintf("java.*%s.*jar", artifactID)).Run()
        }
    }

    time.Sleep(time.Second)
}
