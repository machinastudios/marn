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

# Create dist directory
New-Item -ItemType Directory -Force -Path dist | Out-Null

# Change to src directory
Push-Location src

# Get dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Green
go mod tidy

# Build for Windows amd64
Write-Host "Building for Windows (amd64)..." -ForegroundColor Green
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn-windows-amd64.exe .

# Build for Linux amd64
Write-Host "Building for Linux (amd64)..." -ForegroundColor Green
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn-linux-amd64 .

# Build for Linux ARM64
Write-Host "Building for Linux (arm64)..." -ForegroundColor Green
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn-linux-arm64 .

# Build for macOS amd64
Write-Host "Building for macOS (amd64)..." -ForegroundColor Green
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn-darwin-amd64 .

# Build for macOS ARM64 (M1/M2)
Write-Host "Building for macOS (arm64)..." -ForegroundColor Green
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn-darwin-arm64 .

# Build for current platform
Write-Host "Building for current platform..." -ForegroundColor Green
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
go build -ldflags "-X main.Version=$VERSION" -o ../dist/marn.exe .

Pop-Location

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Binaries are in the dist/ directory:"
Get-ChildItem dist/
