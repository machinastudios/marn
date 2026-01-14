# Marn - Yarn for Maven

Marn is a CLI tool that reads scripts from `pom.xml` and executes them, similar to how `yarn` works with `package.json`.

**Cross-platform:** Works on Windows, Linux, and macOS!

## 📦 Installation

### Quick Install (Recommended)

Download the latest pre-compiled binary for your platform:

[![GitHub Release](https://img.shields.io/github/v/release/machinastudios/marn?style=for-the-badge&logo=github)](https://github.com/machinastudios/marn/releases/latest)

| Platform | Download |
|----------|----------|
| **Windows x64** | [marn-windows-amd64.exe](https://github.com/machinastudios/marn/releases/latest/download/marn-windows-amd64.exe) |
| **Linux x64** | [marn-linux-amd64](https://github.com/machinastudios/marn/releases/latest/download/marn-linux-amd64) |
| **Linux ARM64** | [marn-linux-arm64](https://github.com/machinastudios/marn/releases/latest/download/marn-linux-arm64) |
| **macOS x64** | [marn-darwin-amd64](https://github.com/machinastudios/marn/releases/latest/download/marn-darwin-amd64) |
| **macOS ARM64 (M1/M2)** | [marn-darwin-arm64](https://github.com/machinastudios/marn/releases/latest/download/marn-darwin-arm64) |

> 💡 **Tip:** You can also browse all releases at [github.com/machinastudios/marn/releases](https://github.com/machinastudios/marn/releases)

### Global Installation

After downloading, install globally:

**On Linux/macOS:**

```bash
# Make executable and install
chmod +x marn-linux-amd64
./marn-linux-amd64 init
```

**On Windows (PowerShell):**

```powershell
.\marn-windows-amd64.exe init
```

---

## 🔧 Building from Source

### Prerequisites

- Go 1.21 or higher
- Maven (for running the Maven commands)

### Build

**On Linux/macOS:**

```bash
cd marn
chmod +x build.sh
./build.sh
```

**On Windows (PowerShell):**

```powershell
cd marn
.\build.ps1
```

This will create binaries in the `dist/` directory.

## Usage

After installation, you can use `marn` with the following commands:

| Command | Description |
|---------|-------------|
| `marn init` | Install marn globally (copies binary to PATH) |
| `marn install` | Install dependencies (mvn dependency:resolve) |
| `marn link` | Link current project to local Maven repository (~/.m2) |
| `marn build` | Build the project (mvn clean compile) |
| `marn test` | Run tests (mvn test) |
| `marn package` | Package the project (mvn package) |
| `marn run` | Build and run the JAR |
| `marn clean` | Clean the project (mvn clean) |
| `marn watch` | Watch for changes and rebuild |
| `marn version` | Show version |
| `marn <script>` | Run custom script from pom.xml |

### Linking Projects

If you're working on a local dependency (like `mshared`), use `marn link` to install it to your local Maven repository:

```bash
cd mshared
marn link
```

This runs `mvn clean install -DskipTests`, making the project available to other projects that depend on it.

## Custom Scripts

Define custom scripts in your `pom.xml` under `<properties>`:

```xml
<properties>
    <script.dev>mvn clean compile -DskipTests</script.dev>
    <script.lint>mvn checkstyle:check</script.lint>
</properties>
```

Then run them with:

```bash
marn dev
marn lint
```

### Pre and Post Scripts

You can define `pre*` and `post*` scripts that run before and after any command (including custom scripts):

```xml
<properties>
    <script.preBuild>echo "Preparing build..."</script.preBuild>
    <script.postBuild>cp ${BUILD_ARTIFACT} dist/</script.postBuild>
    
    <script.dev>mvn clean compile</script.dev>
    <script.preDev>echo "Starting dev mode"</script.preDev>
    <script.postDev>echo "Dev mode complete"</script.postDev>
</properties>
```

Pre-scripts run before the main command, and post-scripts run after. If a pre-script fails, the main command won't run. If a post-script fails, the program exits with an error.

## Environment Variables

Marn automatically sets the following environment variables that you can use in your scripts:

### `BUILD_ARTIFACT`
Set after `marn build` and `marn run`. Contains the absolute path to the most recently compiled JAR file.

### `TARGET_DIR`
Set after `marn build`, `marn package`, and `marn run`. Contains the absolute path to the `target/` directory.

### Using Variables in Scripts

You can use environment variables in scripts using `${VAR}` or `$VAR` syntax:

```xml
<properties>
    <script.postBuild>cp ${BUILD_ARTIFACT} ../deploy/</script.postBuild>
    <script.deploy>scp ${BUILD_ARTIFACT} server:/opt/app/</script.deploy>
    <script.copyJars>cp ${TARGET_DIR}/*.jar dist/</script.copyJars>
</properties>
```

## .env File Support

Marn automatically loads environment variables from a `.env` file in your project root if it exists:

```env
DATABASE_URL=jdbc:postgresql://localhost:5432/mydb
API_KEY=secret123
PORT=8080
```

Variables in `.env` won't override existing environment variables. You can use these variables in your scripts:

```xml
<properties>
    <script.run>java -jar ${BUILD_ARTIFACT} --port=${PORT}</script.run>
</properties>
```

## Cross-Platform Commands

Marn provides Unix-like command aliases on Windows, so you can use the same scripts across all platforms:

- `cp` - Copy files/directories (supports wildcards like `*.jar`)
- `mv` - Move/rename files
- `rm` - Remove files
- `mkdir` - Create directories
- `touch` - Create files or update timestamps
- `cat` - Display file contents
- `grep` - Search text in files
- `ls` - List files
- `pwd` - Print working directory
- `which` - Find command location

Example:

```xml
<properties>
    <script.postBuild>cp ${TARGET_DIR}/*.jar ../deploy/</script.postBuild>
    <script.cleanup>rm -rf ${TARGET_DIR}/classes</script.cleanup>
</properties>
```

## Watch Mode

Configure watch mode in your `pom.xml`:

```xml
<properties>
    <watch.dirs>src/main/java src/main/resources</watch.dirs>
    <watch.buildCommand>compile</watch.buildCommand>
    <watch.skipTests>true</watch.skipTests>
    <watch.debounceTime>2</watch.debounceTime>
    <watch.postCommand>./marn run</watch.postCommand>
    <watch.localDeps>../mshared</watch.localDeps>
</properties>
```

Then run:

```bash
marn watch
```

This will watch for changes in the specified directories and rebuild automatically.

## Local Dependencies

Marn automatically detects SNAPSHOT dependencies that have local sibling directories. When you run `marn build`, `marn test`, or `marn watch`, it will:

1. Find dependencies with SNAPSHOT versions
2. Check if a sibling directory with the same artifactId exists
3. Build those dependencies first before building the main project

You can also manually configure local dependencies in `pom.xml`:

```xml
<properties>
    <watch.localDeps>../mshared ../another-dep</watch.localDeps>
</properties>
```

## Project Structure

```
marn/
├── src/                  # Source code
│   ├── main.go           # Entry point and command handling
│   ├── colors.go         # Terminal color definitions
│   ├── colors_windows.go # Windows-specific color handling
│   ├── colors_unix.go    # Unix-specific color handling
│   ├── commands.go       # Main commands (build, test, run, etc.)
│   ├── init.go           # Global installation command
│   ├── pom.go            # POM parsing and property extraction
│   ├── watch.go          # Watch mode implementation
│   ├── utils.go          # Utility functions
│   ├── go.mod            # Go module definition
│   └── go.sum            # Go dependencies checksum
├── dist/                 # Compiled binaries (generated)
├── build.sh              # Linux/macOS build script
├── build.ps1             # Windows build script
├── README.md             # Documentation
└── .github/
    └── workflows/
        └── release.yml   # GitHub Actions workflow
```

## GitHub Actions

The project includes a GitHub Actions workflow that:

1. **On every push to `main`:** Validates that the code builds for all platforms
2. **On tag push (`v*`):** Creates a release with binaries for all platforms

### Creating a Release

To create a new release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow will automatically:
- Build binaries for Linux, Windows, and macOS (both amd64 and arm64)
- Create a GitHub release with all binaries
- Generate release notes
- Include SHA256 checksums

## Platform Differences

### Windows

- Uses PowerShell to run shell commands
- Provides Unix command aliases (`cp`, `mv`, `rm`, etc.)
- Uses `taskkill` to kill existing Java processes
- Installs to `%USERPROFILE%\bin`
- Make sure Maven (`mvn.cmd`) is in your PATH

### Linux/macOS

- Uses `bash -c` to run shell commands
- Uses `pkill` to kill existing Java processes
- Installs to `/usr/local/bin` (may require sudo)
- Uses `fsnotify` for file watching (native inotify on Linux)

## License

MIT
