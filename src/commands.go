package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// installDependencies runs mvn dependency:resolve
func installDependencies() {
	// Run pre-install script
	if err := runPreScript("install"); err != nil {
		fmt.Printf("%s✗ Pre-install script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sInstalling dependencies...%s\n", colors.Green, colors.Reset)

	if err := runMvnCommand("dependency:resolve"); err != nil {
		os.Exit(1)
	}

	// Run post-install script
	if err := runPostScript("install"); err != nil {
		fmt.Printf("%s✗ Post-install script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// linkProject links the project to local Maven repository
func linkProject() {
	if _, err := os.Stat(pomFile); os.IsNotExist(err) {
		fmt.Printf("%sError: pom.xml not found%s\n", colors.Red, colors.Reset)
		fmt.Println("Please run 'marn link' from a Maven project directory.")
		os.Exit(1)
	}

	// Run pre-link script
	if err := runPreScript("link"); err != nil {
		fmt.Printf("%s✗ Pre-link script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sLinking project to local Maven repository...%s\n", colors.Blue, colors.Reset)
	fmt.Println()

	fmt.Printf("%sInstalling project to ~/.m2/repository...%s\n", colors.Green, colors.Reset)

	if err := runMvnCommand("clean", "install", "-DskipTests"); err != nil {
		fmt.Printf("%s✗ Failed to link project%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	// Update hash after successful link
	if err := updateSrcHash(currentDir); err != nil {
		// Log but don't fail the link if hash update fails
		fmt.Printf("%sWarning: Could not update hash: %v%s\n", colors.Yellow, err, colors.Reset)
	}

	fmt.Println()
	fmt.Printf("%s✓ Project linked to local Maven repository!%s\n", colors.Green, colors.Reset)
	fmt.Println()
	fmt.Println("Other projects can now use this as a dependency.")

	// Run post-link script
	if err := runPostScript("link"); err != nil {
		fmt.Printf("%s✗ Post-link script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// buildProject builds the project
func buildProject() {
	// Run pre-build script
	if err := runPreScript("build"); err != nil {
		fmt.Printf("%s✗ Pre-build script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sBuilding project...%s\n", colors.Green, colors.Reset)

	if err := buildLocalDependencies(true); err != nil {
		os.Exit(1)
	}

	// Use package to generate JAR file
	if err := runMvnCommand("clean", "package", "-DskipTests"); err != nil {
		os.Exit(1)
	}

	// Set TARGET_DIR environment variable
	targetDir := filepath.Join(currentDir, "target")
	absTargetDir, err := filepath.Abs(targetDir)
	if err == nil {
		os.Setenv("TARGET_DIR", absTargetDir)
	}

	// Find the most recent JAR and set BUILD_ARTIFACT
	jarFile := findJarFile()
	if jarFile != "" {
		absPath, err := filepath.Abs(jarFile)
		if err == nil {
			os.Setenv("BUILD_ARTIFACT", absPath)
		}
	} else {
		// Clear BUILD_ARTIFACT if no JAR found (mvn compile doesn't create JAR)
		os.Setenv("BUILD_ARTIFACT", "")
	}

	// Run post-build script
	if err := runPostScript("build"); err != nil {
		fmt.Printf("%s✗ Post-build script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// testProject runs tests
func testProject() {
	// Run pre-test script
	if err := runPreScript("test"); err != nil {
		fmt.Printf("%s✗ Pre-test script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sRunning tests...%s\n", colors.Green, colors.Reset)

	if err := buildLocalDependencies(true); err != nil {
		os.Exit(1)
	}

	if err := runMvnCommand("test"); err != nil {
		os.Exit(1)
	}

	// Run post-test script
	if err := runPostScript("test"); err != nil {
		fmt.Printf("%s✗ Post-test script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// packageProject packages the project
func packageProject() {
	// Run pre-package script
	if err := runPreScript("package"); err != nil {
		fmt.Printf("%s✗ Pre-package script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sPackaging project...%s\n", colors.Green, colors.Reset)

	if err := runMvnCommand("clean", "package"); err != nil {
		os.Exit(1)
	}

	// Set TARGET_DIR environment variable
	targetDir := filepath.Join(currentDir, "target")
	absTargetDir, err := filepath.Abs(targetDir)
	if err == nil {
		os.Setenv("TARGET_DIR", absTargetDir)
	}

	// Run post-package script
	if err := runPostScript("package"); err != nil {
		fmt.Printf("%s✗ Post-package script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// cleanProject cleans the project
func cleanProject() {
	// Run pre-clean script
	if err := runPreScript("clean"); err != nil {
		fmt.Printf("%s✗ Pre-clean script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	fmt.Printf("%sCleaning project...%s\n", colors.Green, colors.Reset)

	if err := runMvnCommand("clean"); err != nil {
		os.Exit(1)
	}

	// Run post-clean script
	if err := runPostScript("clean"); err != nil {
		fmt.Printf("%s✗ Post-clean script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
}

// runProject builds and runs the JAR
func runProject() {
	// Run pre-run script
	if err := runPreScript("run"); err != nil {
		fmt.Printf("%s✗ Pre-run script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

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

	// Set TARGET_DIR environment variable
	targetDir := filepath.Join(currentDir, "target")
	absTargetDir, err := filepath.Abs(targetDir)
	if err == nil {
		os.Setenv("TARGET_DIR", absTargetDir)
	}

	// Find the JAR file
	jarFile := findJarFile()
	if jarFile == "" {
		fmt.Printf("%sError: No JAR file found in target/ directory%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}

	// Set BUILD_ARTIFACT
	absPath, err := filepath.Abs(jarFile)
	if err == nil {
		os.Setenv("BUILD_ARTIFACT", absPath)
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

	// Run post-run script
	if err := runPostScript("run"); err != nil {
		fmt.Printf("%s✗ Post-run script failed%s\n", colors.Red, colors.Reset)
		os.Exit(1)
	}
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

	// Run pre-script if it exists
	preScriptName := "pre" + strings.ToUpper(scriptName[:1]) + scriptName[1:]
	if preScript, preExists := scripts[preScriptName]; preExists {
		fmt.Printf("%sRunning pre-script: %s%s\n", colors.Yellow, preScriptName, colors.Reset)
		if err := runShellCommand(preScript); err != nil {
			fmt.Printf("%s✗ Pre-script failed%s\n", colors.Red, colors.Reset)
			os.Exit(1)
		}
	}

	// Expand environment variables in script content
	content = expandEnvVars(content)

	fmt.Printf("%sExecuting script: %s%s\n", colors.Yellow, scriptName, colors.Reset)
	fmt.Printf("%sCommand: %s%s\n", colors.Blue, content, colors.Reset)
	fmt.Println()

	if err := runShellCommand(content); err != nil {
		os.Exit(1)
	}

	// Run post-script if it exists
	postScriptName := "post" + strings.ToUpper(scriptName[:1]) + scriptName[1:]
	if postScript, postExists := scripts[postScriptName]; postExists {
		fmt.Printf("%sRunning post-script: %s%s\n", colors.Yellow, postScriptName, colors.Reset)
		if err := runShellCommand(postScript); err != nil {
			fmt.Printf("%s✗ Post-script failed%s\n", colors.Red, colors.Reset)
			os.Exit(1)
		}
	}
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
			// Check if dependency needs to be rebuilt
			shouldRebuild, currentHash, storedHash, err := shouldRebuildDependency(depPath)
			if err != nil {
				// If we can't check hash, rebuild to be safe
				shouldRebuild = true
			}

			relPath, err := filepath.Rel(currentDir, depPath)
			if err != nil {
				relPath = depPath
			}

			if !shouldRebuild {
				fmt.Printf("%sSkipping dependency (no changes): %s%s\n", colors.Green, relPath, colors.Reset)
				printHashComparison(depPath, currentHash, storedHash)
				continue
			}

			fmt.Printf("%sBuilding dependency: %s%s\n", colors.Blue, relPath, colors.Reset)
			printHashComparison(depPath, currentHash, storedHash)

			args := []string{"clean", "install"}
			if skipTests {
				args = append(args, "-DskipTests")
			}

			cmd := exec.Command("mvn", args...)
			cmd.Dir = depPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("%sFailed to build dependency: %s%s\n", colors.Red, depPath, colors.Reset)
				return err
			}

			// Update hash after successful build
			if err := updateSrcHash(depPath); err != nil {
				// Log but don't fail the build if hash update fails
				fmt.Printf("%sWarning: Could not update hash for %s: %v%s\n", colors.Yellow, depPath, err, colors.Reset)
			}

			fmt.Printf("%s✓ Dependency built: %s%s\n", colors.Green, relPath, colors.Reset)
		}
	}

	fmt.Println()
	return nil
}

// findJarFile finds the most recently modified JAR file in target directory
// Prefers JARs that don't start with "original-"
func findJarFile() string {
	targetDir := filepath.Join(currentDir, "target")

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return ""
	}

	var latestJar string
	var latestTime time.Time
	var latestOriginalJar string
	var latestOriginalTime time.Time

	for _, entry := range entries {

		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jar") {
			jarPath := filepath.Join(targetDir, entry.Name())
			info, err := os.Stat(jarPath)

			if err == nil {
				modTime := info.ModTime()
				isOriginal := strings.HasPrefix(entry.Name(), "original-")

				if isOriginal {
					// Track original JARs separately
					if latestOriginalJar == "" || modTime.After(latestOriginalTime) {
						latestOriginalJar = jarPath
						latestOriginalTime = modTime
					}
				} else {
					// Prefer non-original JARs
					if latestJar == "" || modTime.After(latestTime) {
						latestJar = jarPath
						latestTime = modTime
					}
				}
			}
		}
	}

	// Return non-original JAR if found, otherwise fall back to original JAR
	if latestJar != "" {
		return latestJar
	}

	return latestOriginalJar
}

// runPreScript executes a pre script if it exists
func runPreScript(commandName string) error {
	scripts := getScriptsFromPom()
	preScriptName := "pre" + strings.ToUpper(commandName[:1]) + commandName[1:]

	if script, exists := scripts[preScriptName]; exists {
		fmt.Printf("%sRunning pre-script: %s%s\n", colors.Yellow, preScriptName, colors.Reset)
		return runShellCommand(script)
	}

	return nil
}

// runPostScript executes a post script if it exists
func runPostScript(commandName string) error {
	scripts := getScriptsFromPom()
	postScriptName := "post" + strings.ToUpper(commandName[:1]) + commandName[1:]

	if script, exists := scripts[postScriptName]; exists {
		fmt.Printf("%sRunning post-script: %s%s\n", colors.Yellow, postScriptName, colors.Reset)
		return runShellCommand(script)
	}

	return nil
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
