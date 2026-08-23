param(
    [string]$Version = "",
    [string]$VexVersion = "",
    [switch]$Latest
)

$ErrorActionPreference = "Stop"

$WaveRepo = "wavefnd/Wave"
$VexRepo = "wavefnd/Vex"

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

function Assert-Version($Value) {
    if ($Value -notmatch '^v[0-9A-Za-z][0-9A-Za-z._+-]*$') {
        Fail "Invalid version tag: $Value"
    }
}

function Resolve-LatestVersion($Repository) {
    $releases = Invoke-RestMethod -Headers @{ Accept = "application/vnd.github+json" } -Uri "https://api.github.com/repos/$Repository/releases?per_page=1"
    if ($releases -is [array]) {
        $release = $releases[0]
    } else {
        $release = $releases
    }
    if ($null -eq $release -or [string]::IsNullOrWhiteSpace($release.tag_name)) {
        Fail "Unable to resolve the latest release for $Repository."
    }
    Assert-Version $release.tag_name
    return $release.tag_name
}

function Get-PublishedHash($SumsPath, $FileName) {
    foreach ($line in Get-Content -LiteralPath $SumsPath) {
        if ($line -match '^(?<hash>[0-9A-Fa-f]{64})\s+\*?(?<name>.+)$') {
            if ($Matches['name'].Trim() -eq $FileName) {
                return $Matches['hash'].ToLowerInvariant()
            }
        }
    }
    return ""
}

function Assert-Checksum($ArchivePath, $SumsPath, $FileName) {
    $expected = Get-PublishedHash $SumsPath $FileName
    if ([string]::IsNullOrWhiteSpace($expected)) {
        Fail "No valid checksum was published for $FileName."
    }
    $actual = (Get-FileHash -LiteralPath $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        Fail "Checksum verification failed for $FileName."
    }
    Write-Info "Verified SHA-256: $FileName"
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

function Restore-Installation($InstallDirectory, $BackupDirectory) {
    if (Test-Path -LiteralPath $InstallDirectory) {
        Remove-Item -Recurse -Force -LiteralPath $InstallDirectory -ErrorAction SilentlyContinue
    }
    if (Test-Path -LiteralPath $BackupDirectory) {
        Move-Item -LiteralPath $BackupDirectory -Destination $InstallDirectory
    }
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
    $Version = Resolve-LatestVersion $WaveRepo
    $VexVersion = Resolve-LatestVersion $VexRepo
    Write-Info "Latest Wave version: $Version"
    Write-Info "Latest Vex version: $VexVersion"
} else {
    $Version = Normalize-Version $Version
    $VexVersion = Normalize-Version $VexVersion
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Host "Wave Toolchain Installer"
    Write-Host "Usage:"
    Write-Host "  powershell -ExecutionPolicy Bypass -File install.ps1 -Version <wave-tag> [-VexVersion <vex-tag>]"
    Write-Host "  powershell -ExecutionPolicy Bypass -File install.ps1 -Latest"
    Fail "Missing Wave version. Use -Version <tag> or -Latest."
}
Assert-Version $Version

if ([string]::IsNullOrWhiteSpace($VexVersion)) {
    $VexVersion = Resolve-LatestVersion $VexRepo
    Write-Info "Latest Vex version: $VexVersion"
}
Assert-Version $VexVersion

$waveFileSuffix = "x86_64-pc-windows-gnu"
$vexFileSuffix = "x86_64-pc-windows-msvc"
$waveFileName = "wave-$Version-$waveFileSuffix.zip"
$vexFileName = "vex-$VexVersion-$vexFileSuffix.zip"
$waveUrl = "https://github.com/$WaveRepo/releases/download/$Version/$waveFileName"
$vexUrl = "https://github.com/$VexRepo/releases/download/$VexVersion/$vexFileName"
$waveSumsUrl = "https://github.com/$WaveRepo/releases/download/$Version/SHA256SUMS"
$vexSumsUrl = "https://github.com/$VexRepo/releases/download/$VexVersion/SHA256SUMS"

if ($env:WAVE_INSTALL_DIR) {
    $installDir = [System.IO.Path]::GetFullPath($env:WAVE_INSTALL_DIR)
} else {
    $installDir = Join-Path $env:LOCALAPPDATA "Wave\bin"
}

$installParent = Split-Path -Parent $installDir
New-Item -ItemType Directory -Force -Path $installParent | Out-Null

$tempRoot = Join-Path $installParent (".wave-install-" + [System.Guid]::NewGuid().ToString("N"))
$stageDir = "$installDir.new.$PID"
$backupDir = "$installDir.old.$PID"
$waveDownloadPath = Join-Path $tempRoot $waveFileName
$vexDownloadPath = Join-Path $tempRoot $vexFileName
$waveSumsPath = Join-Path $tempRoot "WAVE_SHA256SUMS"
$vexSumsPath = Join-Path $tempRoot "VEX_SHA256SUMS"
$waveExtractRoot = Join-Path $tempRoot "wave"
$vexExtractRoot = Join-Path $tempRoot "vex"

New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

try {
    Write-Step "[1/4] Downloading Wave $Version and Vex $VexVersion..."
    Write-Info "Download: $waveUrl"
    Invoke-WebRequest -UseBasicParsing -Uri $waveUrl -OutFile $waveDownloadPath
    Write-Info "Download: $vexUrl"
    Invoke-WebRequest -UseBasicParsing -Uri $vexUrl -OutFile $vexDownloadPath
    Invoke-WebRequest -UseBasicParsing -Uri $waveSumsUrl -OutFile $waveSumsPath
    Invoke-WebRequest -UseBasicParsing -Uri $vexSumsUrl -OutFile $vexSumsPath

    Write-Step "[2/4] Verifying release archives..."
    Assert-Checksum $waveDownloadPath $waveSumsPath $waveFileName
    Assert-Checksum $vexDownloadPath $vexSumsPath $vexFileName

    Write-Step "[3/4] Installing Wave toolchain..."
    New-Item -ItemType Directory -Force -Path $waveExtractRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $vexExtractRoot | Out-Null
    Expand-Archive -Force -Path $waveDownloadPath -DestinationPath $waveExtractRoot
    Expand-Archive -Force -Path $vexDownloadPath -DestinationPath $vexExtractRoot

    $wavePackageDir = Join-Path $waveExtractRoot ("wave-$Version-$waveFileSuffix")
    $vexPackageDir = Join-Path $vexExtractRoot ("vex-$VexVersion-$vexFileSuffix")

    if (-not (Test-Path -LiteralPath $wavePackageDir -PathType Container)) {
        Fail "Invalid Wave package layout."
    }
    if (-not (Test-Path -LiteralPath $vexPackageDir -PathType Container)) {
        Fail "Invalid Vex package layout."
    }

    $wavec = Join-Path $wavePackageDir "wavec.exe"
    $llvm = Join-Path $wavePackageDir "llvm"
    $vex = Join-Path $vexPackageDir "vex.exe"
    if (-not (Test-Path -LiteralPath $wavec -PathType Leaf)) {
        Fail "Wave package does not contain wavec.exe."
    }
    if (-not (Test-Path -LiteralPath $llvm -PathType Container)) {
        Fail "Wave package does not contain bundled llvm/."
    }
    if (-not (Test-Path -LiteralPath $vex -PathType Leaf)) {
        Fail "Vex package does not contain vex.exe."
    }

    Remove-Item -Recurse -Force -LiteralPath $stageDir -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force -LiteralPath $backupDir -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path $stageDir | Out-Null

    Copy-Item -Force -Path (Join-Path $wavePackageDir "*") -Destination $stageDir -Recurse
    Copy-Item -Force -LiteralPath $vex -Destination (Join-Path $stageDir "vex.exe")

    $vexNoticeDir = Join-Path $stageDir "share\vex"
    New-Item -ItemType Directory -Force -Path $vexNoticeDir | Out-Null
    foreach ($notice in @("COPYRIGHT", "LICENSE", "NOTICE", "README.md")) {
        $noticePath = Join-Path $vexPackageDir $notice
        if (Test-Path -LiteralPath $noticePath -PathType Leaf) {
            Copy-Item -Force -LiteralPath $noticePath -Destination (Join-Path $vexNoticeDir $notice)
        }
    }

    if (Test-Path -LiteralPath $installDir) {
        Move-Item -LiteralPath $installDir -Destination $backupDir
    }

    try {
        Move-Item -LiteralPath $stageDir -Destination $installDir
    } catch {
        if ((Test-Path -LiteralPath $backupDir) -and -not (Test-Path -LiteralPath $installDir)) {
            Move-Item -LiteralPath $backupDir -Destination $installDir
        }
        throw
    }

    Write-Step "[4/4] Verifying installation..."
    $installedWavec = Join-Path $installDir "wavec.exe"
    $installedVex = Join-Path $installDir "vex.exe"

    try {
        if (-not (Test-Path -LiteralPath $installedWavec -PathType Leaf)) {
            throw "wavec.exe was not found in $installDir."
        }
        if (-not (Test-Path -LiteralPath $installedVex -PathType Leaf)) {
            throw "vex.exe was not found in $installDir."
        }

        & $installedWavec --version
        if ($LASTEXITCODE -ne 0) {
            throw "wavec --version exited with code $LASTEXITCODE"
        }
        & $installedVex --version
        if ($LASTEXITCODE -ne 0) {
            throw "vex --version exited with code $LASTEXITCODE"
        }
    } catch {
        Restore-Installation $installDir $backupDir
        throw
    }

    Remove-Item -Recurse -Force -LiteralPath $backupDir -ErrorAction SilentlyContinue
    Add-UserPath $installDir
    $env:Path = "$installDir;$env:Path"

    Write-Host "Installation completed successfully."
    Write-Info "Installed wavec $Version and vex $VexVersion."
    Write-Host "Restart PowerShell if 'wavec' or 'vex' is not available from PATH."
} finally {
    Remove-Item -Recurse -Force -LiteralPath $tempRoot -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force -LiteralPath $stageDir -ErrorAction SilentlyContinue
}
