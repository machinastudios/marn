package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "os/exec"
    "os/signal"
    "path/filepath"
    "regexp"
    "strings"
    "syscall"
    "time"

    "github.com/fsnotify/fsnotify"
)

// watchMode implements the watch functionality
func watchMode() {

    if _, err := os.Stat(pomFile); os.IsNotExist(err) {
        fmt.Printf("%sError: pom.xml not found%s\n", colors.Red, colors.Reset)
        fmt.Println("Please run 'marn watch' from a Maven project directory.")
        os.Exit(1)
    }

    // Get watch configuration
    config := loadWatchConfig()

    // Get local dependencies
    localDeps := getLocalDependencies()

    // Print configuration
    printWatchBanner(config, localDeps)

    // Build local dependencies first
    if err := buildLocalDependencies(config.SkipTests); err != nil {
        os.Exit(1)
    }

    // Initial build
    fmt.Printf("%sRunning initial build...%s\n", colors.Green, colors.Reset)

    success, _ := runMvnBuild(config.BuildCommand, config.SkipTests)
    if success {
        fmt.Printf("%s✓ Initial build complete!%s\n", colors.Green, colors.Reset)

        if config.PostCommand != "" {
            fmt.Printf("%sRunning post command: %s%s\n", colors.Yellow, config.PostCommand, colors.Reset)
            runShellCommand(config.PostCommand)
        }
    } else {
        fmt.Printf("%s✗ Initial build failed!%s\n", colors.Red, colors.Reset)
        os.Exit(1)
    }

    fmt.Println()

    // Start watching
    startWatcher(config, localDeps)
}

// WatchConfig holds watch mode configuration
type WatchConfig struct {
    WatchDirs    string
    BuildCommand string
    SkipTests    bool
    DebounceTime time.Duration
    PostCommand  string
}

// loadWatchConfig loads watch configuration from pom.xml
func loadWatchConfig() WatchConfig {
    config := WatchConfig{
        WatchDirs:    "src/main/java src/main/resources",
        BuildCommand: "compile",
        SkipTests:    true,
        DebounceTime: 2 * time.Second,
        PostCommand:  "",
    }

    // Override with pom.xml values
    if dirs := getProperty("watch.dirs"); dirs != "" {
        config.WatchDirs = dirs
    }

    if cmd := getProperty("watch.buildCommand"); cmd != "" {
        config.BuildCommand = cmd
    }

    if skipTests := getProperty("watch.skipTests"); skipTests == "false" {
        config.SkipTests = false
    }

    if debounce := getProperty("watch.debounceTime"); debounce != "" {

        if d, err := time.ParseDuration(debounce + "s"); err == nil {
            config.DebounceTime = d
        }
    }

    if post := getProperty("watch.postCommand"); post != "" {
        config.PostCommand = post
    }

    return config
}

// printWatchBanner prints the watch mode banner
func printWatchBanner(config WatchConfig, localDeps []string) {
    fmt.Printf("%s╔════════════════════════════════════════╗%s\n", colors.Blue, colors.Reset)
    fmt.Printf("%s║     Maven Watch Build (Generic)      ║%s\n", colors.Blue, colors.Reset)
    fmt.Printf("%s╚════════════════════════════════════════╝%s\n", colors.Blue, colors.Reset)
    fmt.Println()
    fmt.Printf("%sConfiguration:%s\n", colors.Yellow, colors.Reset)
    fmt.Printf("  %sWatching:%s %s\n", colors.Green, colors.Reset, config.WatchDirs)

    if len(localDeps) > 0 {
        fmt.Printf("  %sLocal Dependencies:%s\n", colors.Green, colors.Reset)

        for _, dep := range localDeps {
            fmt.Printf("    - %s\n", dep)
        }
    }

    fmt.Printf("  %sCommand:%s %s\n", colors.Green, colors.Reset, config.BuildCommand)
    fmt.Printf("  %sSkip Tests:%s %v\n", colors.Green, colors.Reset, config.SkipTests)
    fmt.Printf("  %sDebounce:%s %v\n", colors.Green, colors.Reset, config.DebounceTime)

    if config.PostCommand != "" {
        fmt.Printf("  %sPost Command:%s %s\n", colors.Green, colors.Reset, config.PostCommand)
    }

    fmt.Println()
    fmt.Printf("%sPress Ctrl+C to stop%s\n", colors.Yellow, colors.Reset)
    fmt.Println()
}

// startWatcher starts the file watcher
func startWatcher(config WatchConfig, localDeps []string) {

    // Create watcher
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Printf("%sError: Could not create file watcher: %v%s\n", colors.Red, err, colors.Reset)
        os.Exit(1)
    }
    defer watcher.Close()

    // Add directories to watch
    for _, dir := range strings.Fields(config.WatchDirs) {
        dirPath := filepath.Join(currentDir, dir)

        if _, err := os.Stat(dirPath); err == nil {
            addDirRecursive(watcher, dirPath)
        }
    }

    // Add local dependency directories
    for _, dep := range localDeps {
        srcMain := filepath.Join(dep, "src", "main", "java")
        srcResources := filepath.Join(dep, "src", "main", "resources")

        if _, err := os.Stat(srcMain); err == nil {
            addDirRecursive(watcher, srcMain)
        }

        if _, err := os.Stat(srcResources); err == nil {
            addDirRecursive(watcher, srcResources)
        }
    }

    // Handle Ctrl+C
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Debounce timer
    var lastBuild time.Time

    // Watch loop
    for {
        select {
        case event, ok := <-watcher.Events:

            if !ok {
                return
            }

            // Skip non-relevant events
            if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
                continue
            }

            // Debounce check
            if time.Since(lastBuild) < config.DebounceTime {
                continue
            }

            // Handle the file change
            handleFileChange(event, config, localDeps)
            lastBuild = time.Now()

        case err, ok := <-watcher.Errors:

            if !ok {
                return
            }

            fmt.Printf("%sWatcher error: %v%s\n", colors.Red, err, colors.Reset)

        case <-sigChan:
            fmt.Println()
            fmt.Printf("%sStopping watch mode...%s\n", colors.Yellow, colors.Reset)
            return
        }
    }
}

// handleFileChange handles a file change event
func handleFileChange(event fsnotify.Event, config WatchConfig, localDeps []string) {

    // Check if change is in a local dependency
    isLocalDep := false
    var changedDepPath string

    for _, dep := range localDeps {

        if strings.HasPrefix(event.Name, dep) {
            isLocalDep = true
            changedDepPath = dep
            break
        }
    }

    fmt.Println()
    fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colors.Yellow, colors.Reset)
    fmt.Printf("%sChange detected:%s %s\n", colors.Yellow, colors.Reset, event.Name)
    fmt.Printf("%sEvent:%s %s\n", colors.Yellow, colors.Reset, event.Op.String())

    if isLocalDep {
        // Check if dependency actually needs to be rebuilt
        shouldRebuild, currentHash, storedHash, err := shouldRebuildDependency(changedDepPath)
        if err != nil {
            // If we can't check hash, rebuild to be safe
            shouldRebuild = true
        }

        relPath, err := filepath.Rel(currentDir, changedDepPath)
        if err != nil {
            relPath = changedDepPath
        }

        if !shouldRebuild {
            fmt.Printf("%sSkipping dependency (no changes): %s%s\n", colors.Green, relPath, colors.Reset)
            printHashComparison(changedDepPath, currentHash, storedHash)
        } else {
            fmt.Printf("%sLinking dependency and rebuilding...%s\n", colors.Green, colors.Reset)

            // Build and install the dependency
            fmt.Printf("%sLinking dependency: %s%s\n", colors.Blue, relPath, colors.Reset)
            printHashComparison(changedDepPath, currentHash, storedHash)

            args := []string{"clean", "install"}
            if config.SkipTests {
                args = append(args, "-DskipTests")
            }

            cmd := exec.Command("mvn", args...)
            cmd.Dir = changedDepPath
            cmd.Stdout = io.Discard
            cmd.Stderr = io.Discard

            if err := cmd.Run(); err != nil {
                fmt.Printf("%sFailed to link dependency: %s%s\n", colors.Red, changedDepPath, colors.Reset)
                return
            }

            // Update hash after successful build
            if err := updateSrcHash(changedDepPath); err != nil {
                // Log but don't fail the build if hash update fails
                fmt.Printf("%sWarning: Could not update hash for %s: %v%s\n", colors.Yellow, changedDepPath, err, colors.Reset)
            }

            fmt.Printf("%s✓ Dependency linked: %s%s\n", colors.Green, relPath, colors.Reset)
        }
    }

    // Rebuild project
    fmt.Printf("%sRebuilding...%s\n", colors.Green, colors.Reset)

    success, _ := runMvnBuild(config.BuildCommand, config.SkipTests)
    if success {

        if isLocalDep {
            fmt.Printf("%s✓ Dependency linked and build successful!%s\n", colors.Green, colors.Reset)
        } else {
            fmt.Printf("%s✓ Build successful!%s\n", colors.Green, colors.Reset)
        }

        if config.PostCommand != "" {
            fmt.Printf("%sRunning post command: %s%s\n", colors.Yellow, config.PostCommand, colors.Reset)
            runShellCommand(config.PostCommand)
        }
    } else {
        fmt.Printf("%s✗ Build failed!%s\n", colors.Red, colors.Reset)
    }

    fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colors.Yellow, colors.Reset)
    fmt.Println()
}

// runMvnBuild runs Maven build and captures output
func runMvnBuild(command string, skipTests bool) (bool, error) {
    mvnCmd := getMvnCommand()

    args := strings.Fields(command)
    if skipTests {
        args = append(args, "-DskipTests")
    }

    cmd := exec.Command(mvnCmd, args...)
    cmd.Dir = currentDir

    // Capture output
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()

    if err := cmd.Start(); err != nil {
        return false, err
    }

    // Read and filter output
    go filterOutput(stdout)
    go filterOutput(stderr)

    err := cmd.Wait()
    return err == nil, err
}

// filterOutput filters Maven output to show only relevant lines
func filterOutput(r io.Reader) {
    scanner := bufio.NewScanner(r)
    re := regexp.MustCompile(`(ERROR|BUILD|Compiling|SUCCESS|FAILURE)`)

    for scanner.Scan() {
        line := scanner.Text()

        if re.MatchString(line) {
            fmt.Println(line)
        }
    }
}

// addDirRecursive adds a directory and all subdirectories to the watcher
func addDirRecursive(watcher *fsnotify.Watcher, dir string) error {
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

        if err != nil {
            return err
        }

        if info.IsDir() {
            return watcher.Add(path)
        }

        return nil
    })
}
