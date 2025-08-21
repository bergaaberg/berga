# Berga CLI Installation Script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/bergaaberg/berga/main/install/install.ps1 | iex

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\berga\bin",
    [switch]$AddToPath = $true
)

$ErrorActionPreference = "Stop"

# Configuration
$REPO = "bergaaberg/berga"
$BINARY_NAME = "berga.exe"

# Helper functions
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    
    $colorMap = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "White" = [ConsoleColor]::White
    }
    
    Write-Host $Message -ForegroundColor $colorMap[$Color]
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
    exit 1
}

function Test-Admin {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        return "amd64"
    } else {
        return "386"
    }
}

function Get-LatestVersion {
    Write-Info "Fetching latest release information..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
        $version = $response.tag_name
        Write-Info "Latest version: $version"
        return $version
    }
    catch {
        Write-Error "Failed to get latest release version: $_"
    }
}

function Download-Binary {
    param([string]$Version, [string]$Architecture)
    
    $binaryFile = "berga-windows-$Architecture.exe"
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/$binaryFile"
    $tempFile = Join-Path $env:TEMP $binaryFile
    
    Write-Info "Downloading from: $downloadUrl"
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
        Write-Info "Downloaded to: $tempFile"
        return $tempFile
    }
    catch {
        Write-Error "Failed to download binary: $_"
    }
}

function Install-Binary {
    param([string]$SourcePath, [string]$InstallDir)
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    $destinationPath = Join-Path $InstallDir $BINARY_NAME
    
    # Copy binary to install directory
    Write-Info "Installing to: $destinationPath"
    Copy-Item $SourcePath $destinationPath -Force
    
    # Clean up temp file
    Remove-Item $SourcePath -Force
    
    Write-Success "Binary installed to: $destinationPath"
    return $destinationPath
}

function Add-ToPath {
    param([string]$InstallDir)
    
    Write-Info "Adding $InstallDir to PATH..."
    
    # Get current user PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
    
    # Check if already in PATH
    if ($currentPath -like "*$InstallDir*") {
        Write-Info "Install directory already in PATH"
        return
    }
    
    # Add to PATH
    $newPath = "$currentPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
    
    # Update current session PATH
    $env:PATH = "$env:PATH;$InstallDir"
    
    Write-Success "Added to PATH. You may need to restart your terminal for changes to take effect."
}

function Test-Installation {
    param([string]$BinaryPath)
    
    try {
        $version = & $BinaryPath --version 2>$null
        if ($version) {
            Write-Success "Installation verified! Version: $version"
        } else {
            Write-Success "Installation verified!"
        }
        
        Write-Info "Run 'berga --help' to get started"
        Write-Info "Run 'berga config init' to initialize your configuration"
        return $true
    }
    catch {
        Write-Warning "Binary installed but verification failed. Try running: $BinaryPath --help"
        return $false
    }
}

# Main installation flow
function Main {
    Write-Info "Starting Berga CLI installation for Windows..."
    
    $architecture = Get-Architecture
    Write-Info "Detected architecture: $architecture"
    
    $version = Get-LatestVersion
    $tempBinary = Download-Binary $version $architecture
    $installedPath = Install-Binary $tempBinary $InstallDir
    
    if ($AddToPath) {
        Add-ToPath $InstallDir
    }
    
    Test-Installation $installedPath
    
    Write-Success "Installation complete!"
    Write-Host ""
    Write-Host "Quick start:"
    Write-Host "  berga config init    # Initialize configuration"
    Write-Host "  berga script list    # List available scripts"  
    Write-Host "  berga --help         # Show help"
    
    if ($AddToPath) {
        Write-Host ""
        Write-Warning "If 'berga' command is not found, restart your terminal or run:"
        Write-Host "  `$env:PATH += ';$InstallDir'"
    }
}

# Run main function
Main
