# Build script for marn - builds for Windows, Linux, and macOS (PowerShell)

# Get version from git tag or use dev
try {
    $VERSION = git describe --tags 2>$null
    if (-not $VERSION) {
        $SHORT_HASH = git rev-parse --short HEAD 2>$null
        if ($SHORT_HASH) {
            $VERSION = "dev-$SHORT_HASH"
        } else {
            $VERSION = "dev-local"
        }
    }
} catch {
    $VERSION = "dev-local"
}

Write-Host "Building marn..." -ForegroundColor Blue
Write-Host "Version: $VERSION" -ForegroundColor Yellow
Write-Host ""

# Create dist directories
New-Item -ItemType Directory -Force -Path dist | Out-Null
New-Item -ItemType Directory -Force -Path dist/linux | Out-Null
New-Item -ItemType Directory -Force -Path dist/macos | Out-Null

# Change to src directory
Push-Location src

# Get dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Green
go mod tidy

# Build for Windows amd64
Write-Host "Building for Windows (amd64)..." -ForegroundColor Green
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn.exe .

# Build for Linux amd64
Write-Host "Building for Linux (amd64)..." -ForegroundColor Green
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/linux/marn .

# Build for macOS amd64
Write-Host "Building for macOS (amd64)..." -ForegroundColor Green
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/macos/marn .

Pop-Location

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Binaries are in the dist/ directory:"
Get-ChildItem dist/
