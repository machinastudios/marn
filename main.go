package main

import (
    "fmt"
    "os"
    "path/filepath"
)

// Version is set at build time via ldflags
var Version = "dev"

// Global variables
var (
    currentDir string
    pomFile    string
)

func main() {
    var err error

    // Get current working directory
    currentDir, err = os.Getwd()
    if err != nil {
        fmt.Printf("%sError: Could not get current directory%s\n", colors.Red, colors.Reset)
        os.Exit(1)
    }

    pomFile = filepath.Join(currentDir, "pom.xml")

    // Check if pom.xml exists
    if _, err := os.Stat(pomFile); os.IsNotExist(err) {

        // If pom.xml doesn't exist, check if we're being called with init, version or help
        if len(os.Args) < 2 || (os.Args[1] != "init" && os.Args[1] != "--help" && os.Args[1] != "-h" && os.Args[1] != "help" && os.Args[1] != "--version" && os.Args[1] != "-v" && os.Args[1] != "version") {
            fmt.Printf("%sError: pom.xml not found%s\n", colors.Red, colors.Reset)
            fmt.Println("Please run 'marn' commands from a Maven project directory, or")
            fmt.Println("run 'marn init' to install marn globally.")
            os.Exit(1)
        }
    }

    // Handle commands
    if len(os.Args) < 2 {
        showHelp()
        return
    }

    command := os.Args[1]

    switch command {
    case "init":
        initMarn()
    case "install":
        installDependencies()
    case "link":
        linkProject()
    case "install-deps":
        installDependencies()
    case "build":
        buildProject()
    case "test":
        testProject()
    case "package":
        packageProject()
    case "run":
        runProject()
    case "clean":
        cleanProject()
    case "watch":
        watchMode()
    case "version", "--version", "-v":
        showVersion()
    case "help", "--help", "-h":
        showHelp()
    default:
        // Check if it's a custom script
        executeScript(command)
    }
}

// showVersion displays the version
func showVersion() {
    fmt.Printf("marn version %s\n", Version)
}

// showHelp displays the help message
func showHelp() {
    fmt.Printf("%sMarn - Yarn for Maven%s\n", colors.Blue, colors.Reset)
    fmt.Printf("Version: %s\n", Version)
    fmt.Println()
    fmt.Println("Usage: marn <command>")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  init         Install marn globally (copies binary to PATH)")
    fmt.Println("  install      Install dependencies (mvn dependency:resolve)")
    fmt.Println("  link         Link current project to local Maven repository (~/.m2)")
    fmt.Println("  install-deps Install dependencies (mvn dependency:resolve)")
    fmt.Println("  build        Build the project (mvn clean compile)")
    fmt.Println("  test         Run tests (mvn test)")
    fmt.Println("  package      Package the project (mvn package)")
    fmt.Println("  run          Build and run the JAR")
    fmt.Println("  clean        Clean the project (mvn clean)")
    fmt.Println("  watch        Watch for changes and rebuild")
    fmt.Println("  version      Show version")
    fmt.Println("  <script>     Run custom script from pom.xml")
    fmt.Println()
    fmt.Println("Custom scripts are defined in pom.xml under <properties>:")
    fmt.Println("  <script.myScript>mvn compile</script.myScript>")
    fmt.Println()

    if _, err := os.Stat(pomFile); err == nil {
        listScripts()
    }
}

// listScripts lists all available scripts from pom.xml
func listScripts() {
    scripts := getScriptsFromPom()

    if len(scripts) > 0 {
        fmt.Printf("%sAvailable scripts:%s\n", colors.Blue, colors.Reset)

        for name := range scripts {
            fmt.Printf("  %s%s%s\n", colors.Green, name, colors.Reset)
        }
    }
}
