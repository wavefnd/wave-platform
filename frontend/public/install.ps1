param(
    [switch]$Latest,
    [switch]$WithVex,
    [switch]$WithoutVex,
    [switch]$NoModifyPath,
    [Alias('WaveVersion')][string]$Version,
    [string]$VexVersion,
    [switch]$Help,
    [Parameter(ValueFromRemainingArguments = $true)][string[]]$Remaining
)
# Wave installer channel policy: pinned-0.2.1-v1
$ErrorActionPreference = 'Stop'

function Write-Info($Message) { Write-Host "[info] $Message" }
function Manual-Only {
    throw 'This installer installs Wave v0.2.1-pre-beta. Install older versions or Nightly manually: https://github.com/wavefnd/Wave/releases'
}
function Get-LatestRelease($Repository) {
    $best = $null
    for ($page = 1; ; $page++) {
        $result = Invoke-RestMethod -Headers @{Accept='application/vnd.github+json'} -TimeoutSec 60 `
            -Uri "https://api.github.com/repos/$Repository/releases?per_page=100&page=$page"
        $response = @($result)
        foreach ($release in $response) {
            if ($null -eq $release -or $release.draft -ne $false -or -not $release.published_at -or
                $release.tag_name -cnotmatch '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$') { continue }
            if ($null -eq $best -or [DateTimeOffset]$release.published_at -gt [DateTimeOffset]$best.published_at -or
                ([DateTimeOffset]$release.published_at -eq [DateTimeOffset]$best.published_at -and $release.id -gt $best.id)) {
                $best = $release
            }
        }
        if ($response.Count -lt 100) { break }
    }
    if ($null -eq $best) { throw "No public versioned release is available for $Repository." }
    return $best
}
function Get-PinnedWaveRelease {
    $release = Invoke-RestMethod -Headers @{Accept='application/vnd.github+json'} -TimeoutSec 60 -Uri 'https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta'
    if ($release.tag_name -ne 'v0.2.1-pre-beta' -or $release.draft -ne $false) { throw 'Pinned Wave release metadata does not match.' }
    return $release
}
function Get-ReleaseAsset($Release, $Name, [bool]$Optional = $false) {
    $assets = @($Release.assets | Where-Object { $_.name -ceq $Name })
    if ($assets.Count -eq 0 -and $Optional) { return $null }
    if ($assets.Count -ne 1) { throw "Latest release has no unique package: $Name. No older version will be selected." }
    $asset = $assets[0]
    if ($asset.state -ne 'uploaded' -or $asset.digest -cnotmatch '^sha256:[0-9A-Fa-f]{64}$') {
        throw "No valid GitHub SHA-256 was published for $Name."
    }
    return $asset
}
function Assert-Hash($Path, $Expected) {
    if ($Expected -notmatch '^[0-9A-Fa-f]{64}$') { throw "Invalid SHA-256 for $Path." }
    $actual = (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
    if ($actual -ine $Expected) { throw "SHA-256 verification failed: $Path" }
}
function Get-WindowsTarget($Architecture) {
    switch ($Architecture.ToLowerInvariant()) {
        { $_ -in @('x64','amd64','x86_64') } { return 'x86_64-pc-windows-msvc' }
        { $_ -in @('arm64','aarch64') } { return 'aarch64-pc-windows-msvc' }
        default { throw "Unsupported Windows architecture: $Architecture" }
    }
}
function Get-NativeWindowsTarget {
    if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) { throw 'Use install.sh on Linux, macOS or FreeBSD.' }
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    try { $architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString() }
    catch { $architecture = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE } }
    return Get-WindowsTarget $architecture
}
function Save-Download($Uri, $Path) {
    if (-not $Uri.StartsWith('https://github.com/')) { throw 'Expected a GitHub HTTPS download.' }
    for ($attempt = 1; $attempt -le 3; $attempt++) {
        try {
            Invoke-WebRequest -UseBasicParsing -Uri $Uri -OutFile $Path -TimeoutSec 600
            return
        } catch {
            if ($attempt -eq 3) { throw }
            Start-Sleep -Seconds 2
        }
    }
}
function Add-UserPath($Directory) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @($current -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if (-not @($parts | Where-Object { $_.TrimEnd('\') -ieq $Directory.TrimEnd('\') }).Count) {
        [Environment]::SetEnvironmentVariable('Path', (($parts + $Directory) -join ';'), 'User')
    }
    Write-Info 'User PATH configured. Open a new terminal to use Wave.'
}
function Test-Installation($Directory, $Target, [bool]$InstallVex, $Work) {
    $wavec = Join-Path $Directory 'wavec.exe'
    & $wavec --version
    if ($LASTEXITCODE -ne 0) { throw 'wavec --version failed. Check the Visual C++ runtime.' }
    $spec = & $wavec print target-spec --format=json
    $actual = ($spec | ConvertFrom-Json).triple
    if ($LASTEXITCODE -ne 0 -or "$actual".Trim() -ne $Target) { throw "Compiler host target differs from $Target." }
    $source = Join-Path $Work 'install-smoke.wave'
    [IO.File]::WriteAllText($source, @'
import("std::mem::layout")::{size_of};
fun main() -> i32 {
    if (size_of<i64>() != 8) { return 1; }
    return 0;
}
'@)
    Push-Location $Work
    try {
        & $wavec run $source --std-root (Join-Path $Directory 'std')
        if ($LASTEXITCODE -ne 0) {
            throw 'Bundled std compile/run check failed. Ensure the Windows SDK and MSVC/UCRT libraries are installed and available (for example, in a Visual Studio developer shell).'
        }
    } finally { Pop-Location }
    if ($InstallVex) {
        & (Join-Path $Directory 'vex.exe') --version
        if ($LASTEXITCODE -ne 0) { throw 'vex --version failed.' }
    }
}
function Install-Wave {
    if ($Help) {
        Write-Host @'
Wave Toolchain Installer — Wave v0.2.1-pre-beta
Usage: .\install.ps1 [-Latest] [-WithVex | -WithoutVex] [-NoModifyPath]
Default: install Wave and install Vex when its latest release supports this platform.
-WithVex requires Vex; -WithoutVex installs Wave only.
WAVE_INSTALL_DIR overrides the dedicated installation directory (%LOCALAPPDATA%\Wave\bin).
Older versions and Nightly: download manually from https://github.com/wavefnd/Wave/releases
'@
        return
    }
    if ($Version -or $VexVersion -or $env:WAVE_VERSION -or $env:VEX_VERSION) { Manual-Only }
    if ($Remaining.Count) {
        if (@($Remaining | Where-Object { $_ -match '(?i)nightly|version|^v?\d' }).Count) { Manual-Only }
        throw "Unknown argument: $($Remaining -join ' '). Use -Help."
    }
    if ($WithVex -and $WithoutVex) { throw 'Conflicting Vex options.' }
    $target = Get-NativeWindowsTarget
    $installDir = if ($env:WAVE_INSTALL_DIR) { [IO.Path]::GetFullPath($env:WAVE_INSTALL_DIR) } else { Join-Path $env:LOCALAPPDATA 'Wave\bin' }
    if ($installDir -match '[;\r\n]') { throw 'Use an installation directory without semicolons or newlines.' }
    if ($installDir.TrimEnd('\') -eq [IO.Path]::GetPathRoot($installDir).TrimEnd('\') -or
        $installDir.TrimEnd('\') -ieq $HOME.TrimEnd('\')) { throw 'Use a dedicated installation directory.' }
    if (Test-Path -LiteralPath $installDir) {
        if ((Get-Item -LiteralPath $installDir).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'The installation directory must not be a link or junction.' }
        if (-not (Test-Path -LiteralPath (Join-Path $installDir 'wavec.exe') -PathType Leaf)) {
            throw "Refusing to replace a directory not managed by Wave: $installDir"
        }
    }
    $wave = Get-PinnedWaveRelease
    $waveName = "wave-$($wave.tag_name)-$target.zip"
    $waveAsset = Get-ReleaseAsset $wave $waveName
    $vex = $null; $vexAsset = $null; $vexName = ''
    if (-not $WithoutVex) {
        $vex = Get-LatestRelease 'wavefnd/Vex'
        $vexName = "vex-$($vex.tag_name)-$target.zip"
        $vexAsset = Get-ReleaseAsset $vex $vexName $true
        if ($null -eq $vexAsset) {
            if ($WithVex) { throw "Latest Vex has no package for $target." }
            Write-Info "Latest Vex has no package for $target; installing Wave only."
        }
    }
    Write-Info "Wave $($wave.tag_name) / $target"
    Write-Info "Install directory: $installDir"
    $parent = Split-Path -Parent $installDir
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    $lockPath = "$installDir.install-lock"
    # CreateNew prevents simultaneous installers from moving the same installation.
    $lock = [IO.File]::Open($lockPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
    $temp = Join-Path $parent ('.wave-install-' + [Guid]::NewGuid().ToString('N'))
    $stage = Join-Path $temp 'stage'; $backup = Join-Path $temp 'previous'
    $activated = $false; $committed = $false; $keepBackup = $false
    try {
        New-Item -ItemType Directory -Path $temp | Out-Null
        Write-Info '[1/4] Downloading and verifying packages'
        $waveZip = Join-Path $temp 'wave.zip'
        Save-Download "https://github.com/wavefnd/Wave/releases/download/$($wave.tag_name)/$waveName" $waveZip
        Assert-Hash $waveZip $waveAsset.digest.Substring(7)
        if ($vexAsset) {
            $vexZip = Join-Path $temp 'vex.zip'
            Save-Download "https://github.com/wavefnd/Vex/releases/download/$($vex.tag_name)/$vexName" $vexZip
            Assert-Hash $vexZip $vexAsset.digest.Substring(7)
        }
        Write-Info '[2/4] Preparing installation'
        $waveRoot = Join-Path $temp 'wave'
        Expand-Archive -LiteralPath $waveZip -DestinationPath $waveRoot
        $package = Join-Path $waveRoot ([IO.Path]::GetFileNameWithoutExtension($waveName))
        foreach ($path in @('wavec.exe','llvm\bin','std\manifest.json')) {
            if (-not (Test-Path -LiteralPath (Join-Path $package $path))) { throw "Wave package is missing $path." }
        }
        New-Item -ItemType Directory -Path $stage | Out-Null
        Get-ChildItem -LiteralPath $package -Force | Copy-Item -Destination $stage -Recurse -Force
        if ($vexAsset) {
            $vexRoot = Join-Path $temp 'vex'
            Expand-Archive -LiteralPath $vexZip -DestinationPath $vexRoot
            $package = Join-Path $vexRoot ([IO.Path]::GetFileNameWithoutExtension($vexName))
            Copy-Item -LiteralPath (Join-Path $package 'vex.exe') -Destination $stage
            $notices = Join-Path $stage 'share\vex'
            New-Item -ItemType Directory -Force -Path $notices | Out-Null
            foreach ($name in @('COPYRIGHT','LICENSE','NOTICE','README.md')) {
                $path = Join-Path $package $name
                if (Test-Path -LiteralPath $path) { Copy-Item -LiteralPath $path -Destination $notices }
            }
        }
        Write-Info '[3/4] Activating installation'
        if (Test-Path -LiteralPath $installDir) { Move-Item -LiteralPath $installDir -Destination $backup }
        Move-Item -LiteralPath $stage -Destination $installDir
        $activated = $true
        Write-Info '[4/4] Checking compiler, bundled std and runtime'
        Test-Installation $installDir $target ($null -ne $vexAsset) $temp
        $committed = $true
        if (-not $NoModifyPath) {
            try { Add-UserPath $installDir }
            catch { Write-Warning "Wave is installed, but PATH could not be configured. Add $installDir to PATH manually." }
        }
        Write-Info "Installed Wave $($wave.tag_name)."
        if ($vexAsset) { Write-Info "Installed Vex $($vex.tag_name)." }
    } finally {
        try {
            if (-not $committed) {
                if ($activated) { Remove-Item -LiteralPath $installDir -Recurse -Force }
                if (Test-Path -LiteralPath $backup) {
                    try { Move-Item -LiteralPath $backup -Destination $installDir }
                    catch { $keepBackup = $true; throw "Restore failed. Previous installation retained at $backup. $($_.Exception.Message)" }
                }
            }
            if (-not $keepBackup -and (Test-Path -LiteralPath $temp)) { Remove-Item -LiteralPath $temp -Recurse -Force }
        } finally {
            $lock.Dispose()
            Remove-Item -LiteralPath $lockPath -Force
        }
    }
}
# Dot-sourcing exposes helpers without installing anything.
if ($MyInvocation.InvocationName -ne '.') {
    try { Install-Wave }
    catch { Write-Error $_; exit 1 }
}
