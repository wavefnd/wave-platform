param(
    [string]$Version = "",
    [switch]$Latest
)

$ErrorActionPreference = "Stop"

$Repo = "wavefnd/Wave"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Write-Info($Message) {
    Write-Host "[info] $Message"
}

function Write-Step($Message) {
    Write-Host $Message
}

function Fail($Message) {
    Write-Error "[error] $Message"
    exit 1
}

function Normalize-Version($Value) {
    if ([string]::IsNullOrWhiteSpace($Value)) {
        return ""
    }
    if ($Value.StartsWith("v")) {
        return $Value
    }
    return "v$Value"
}

function Add-UserPath($Directory) {
    $fullPath = [System.IO.Path]::GetFullPath($Directory)
    $current = [Environment]::GetEnvironmentVariable("Path", "User")

    if ([string]::IsNullOrWhiteSpace($current)) {
        $next = $fullPath
    } else {
        $parts = $current -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
        if ($parts -contains $fullPath) {
            return
        }
        $next = ($parts + $fullPath) -join ';'
    }

    [Environment]::SetEnvironmentVariable("Path", $next, "User")
    Write-Info "Added $fullPath to user PATH"
}

Write-Info "Detecting system..."

if ([System.Environment]::OSVersion.Platform -ne [System.PlatformID]::Win32NT) {
    Fail "install.ps1 is for Windows. Use install.sh on Linux/macOS."
}

$arch = $env:PROCESSOR_ARCHITECTURE
if (-not [Environment]::Is64BitOperatingSystem -or ($arch -ne "AMD64" -and $arch -ne "x86_64")) {
    Fail "Windows installer currently supports x86_64 only."
}

if ($Latest) {
    $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases?per_page=1"
    if ($release -is [array]) {
        $release = $release[0]
    }
    $Version = $release.tag_name
    Write-Info "Latest version: $Version"
} else {
    $Version = Normalize-Version $Version
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Host "Wave Installer"
    Write-Host "Usage:"
    Write-Host "  powershell -ExecutionPolicy Bypass -File install.ps1 -Version <tag>"
    Write-Host "  powershell -ExecutionPolicy Bypass -File install.ps1 -Latest"
    Fail "Missing version. Use -Version <tag> or -Latest."
}

Write-Step "[1/3] Downloading Wave $Version..."

$fileSuffix = "x86_64-pc-windows-gnu"
$fileName = "wave-$Version-$fileSuffix.zip"
$url = "https://github.com/$Repo/releases/download/$Version/$fileName"
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("wave-install-" + [System.Guid]::NewGuid().ToString("N"))
$downloadPath = Join-Path $tempRoot $fileName

New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

try {
    Write-Info "Download: $url"
    Invoke-WebRequest -Uri $url -OutFile $downloadPath

    Write-Step "[2/3] Installing Wave..."

    Expand-Archive -Force -Path $downloadPath -DestinationPath $tempRoot

    $packageDir = Join-Path $tempRoot ("wave-$Version-$fileSuffix")
    if (-not (Test-Path $packageDir)) {
        $packageDir = Get-ChildItem -Path $tempRoot -Directory | Select-Object -First 1 -ExpandProperty FullName
    }

    if (-not $packageDir -or -not (Test-Path $packageDir)) {
        Fail "Invalid package layout."
    }

    $wavec = Join-Path $packageDir "wavec.exe"
    $llvm = Join-Path $packageDir "llvm"
    if (-not (Test-Path $wavec)) {
        Fail "Package does not contain wavec.exe."
    }
    if (-not (Test-Path $llvm)) {
        Fail "Package does not contain bundled llvm/."
    }

    if ($env:WAVE_INSTALL_DIR) {
        $installDir = $env:WAVE_INSTALL_DIR
    } else {
        $installDir = Join-Path $env:LOCALAPPDATA "Wave\bin"
    }

    New-Item -ItemType Directory -Force -Path $installDir | Out-Null

    $installedLlvm = Join-Path $installDir "llvm"
    if (Test-Path $installedLlvm) {
        Remove-Item -Recurse -Force $installedLlvm
    }

    Copy-Item -Force -Path (Join-Path $packageDir "*") -Destination $installDir -Recurse
    Add-UserPath $installDir
    $env:Path = "$installDir;$env:Path"

    Write-Step "[3/3] Verifying installation..."

    $installedWavec = Join-Path $installDir "wavec.exe"
    if (-not (Test-Path $installedWavec)) {
        Fail "Installation failed. wavec.exe was not found in $installDir."
    }

    & $installedWavec --version
    Write-Host "Installation completed successfully."
    Write-Host "Restart PowerShell if 'wavec' is not available from PATH."
} finally {
    Remove-Item -Recurse -Force $tempRoot -ErrorAction SilentlyContinue
}
