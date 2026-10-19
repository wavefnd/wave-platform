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
function Get-LegacyStdRoot {
    $profileRoot = if (-not [string]::IsNullOrEmpty($env:HOME)) { $env:HOME } else { $env:USERPROFILE }
    if ([string]::IsNullOrEmpty($profileRoot)) { throw 'HOME or USERPROFILE is required to locate the compiler standard library.' }
    $profilePathRoot = [IO.Path]::GetPathRoot($profileRoot)
    if (-not [IO.Path]::IsPathRooted($profileRoot) -or
        ([Environment]::OSVersion.Platform -eq [PlatformID]::Win32NT -and $profilePathRoot.Length -lt 3)) {
        throw 'HOME or USERPROFILE must be an absolute directory path.'
    }
    return [IO.Path]::GetFullPath((Join-Path $profileRoot '.wave/lib/wave/std'))
}
function Resolve-InstallerPath($Path, [int]$Depth = 0) {
    if ($Depth -gt 32) { throw 'Too many links while resolving installation paths.' }
    if ([Environment]::OSVersion.Platform -eq [PlatformID]::Win32NT -and
        ($Path.StartsWith('\\?\') -or $Path.StartsWith('\\.\'))) {
        throw 'Use ordinary filesystem paths instead of Windows device paths for installation.'
    }
    $full = [IO.Path]::GetFullPath($Path)
    $root = [IO.Path]::GetPathRoot($full)
    $resolved = $root
    $parts = $full.Substring($root.Length).Split([char[]]'\/', [StringSplitOptions]::RemoveEmptyEntries)
    foreach ($part in $parts) {
        $candidate = Join-Path $resolved $part
        $item = Get-Item -LiteralPath $candidate -Force -ErrorAction SilentlyContinue
        if ($null -ne $item -and ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
            $targets = @($item.Target)
            if ($targets.Count -ne 1 -or [string]::IsNullOrEmpty($targets[0])) {
                throw "Cannot safely resolve installation link: $candidate"
            }
            $target = $targets[0]
            if (-not [IO.Path]::IsPathRooted($target)) { $target = Join-Path (Split-Path -Parent $candidate) $target }
            $resolved = Resolve-InstallerPath $target ($Depth + 1)
        } else {
            $resolved = $candidate
        }
    }
    $trimmed = $resolved.TrimEnd([char[]]'\/')
    $resolvedRoot = [IO.Path]::GetPathRoot($resolved)
    if ($trimmed -eq $resolvedRoot.TrimEnd([char[]]'\/')) { return $resolvedRoot }
    return $trimmed
}
function Assert-SeparateInstallationPaths($InstallDirectory, $StdDirectory) {
    $compilerPath = Resolve-InstallerPath $InstallDirectory
    $stdPath = Resolve-InstallerPath $StdDirectory
    $separator = [IO.Path]::DirectorySeparatorChar
    $comparison = [StringComparison]::OrdinalIgnoreCase
    if ($compilerPath.Equals($stdPath, $comparison) -or
        $compilerPath.StartsWith($stdPath.TrimEnd([char[]]'\/') + $separator, $comparison) -or
        $stdPath.StartsWith($compilerPath.TrimEnd([char[]]'\/') + $separator, $comparison)) {
        throw 'The compiler installation and legacy std directory must not overlap.'
    }
}
function Assert-ManagedStdDirectory($Directory) {
    $item = Get-Item -LiteralPath $Directory -Force -ErrorAction SilentlyContinue
    if ($null -eq $item) { return }
    if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'The legacy std directory must not be a link or junction.' }
    if (-not $item.PSIsContainer -or -not (Test-Path -LiteralPath (Join-Path $Directory 'manifest.json') -PathType Leaf)) {
        throw "Refusing to replace a standard library not managed by Wave: $Directory"
    }
}
function Invoke-InstalledTool($Executable, [string[]]$Arguments) {
    $output = & $Executable @Arguments
    if ($LASTEXITCODE -ne 0) { throw "Installed tool failed ($LASTEXITCODE): $Executable $($Arguments -join ' ')" }
    return $output
}
function Test-Installation($Directory, $Target, [bool]$InstallVex, $Work) {
    $wavec = Join-Path $Directory 'wavec.exe'
    Invoke-InstalledTool $wavec @('--version') | Write-Host
    $spec = Invoke-InstalledTool $wavec @('print', 'target-spec', '--format=json')
    $actual = ($spec | ConvertFrom-Json).triple
    if ("$actual".Trim() -ne $Target) { throw "Compiler host target differs from $Target." }
    $stdPath = Invoke-InstalledTool $wavec @('print', 'std-path')
    $expectedStd = Resolve-InstallerPath (Get-LegacyStdRoot)
    if (-not (Resolve-InstallerPath ("$stdPath".Trim())).Equals($expectedStd, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Compiler default std path differs from the installed standard library.'
    }
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
        Invoke-InstalledTool $wavec @('run', $source) | Write-Host
    } catch {
        throw "Default std compile/run check failed. Ensure the Windows SDK and MSVC/UCRT libraries are installed and available (for example, in a Visual Studio developer shell). $($_.Exception.Message)"
    } finally { Pop-Location }
    if ($InstallVex) { Invoke-InstalledTool (Join-Path $Directory 'vex.exe') @('--version') | Write-Host }
}
function Install-Wave {
    if ($Help) {
        Write-Host @'
Wave Toolchain Installer — Wave v0.2.1-pre-beta
Usage: .\install.ps1 [-Latest] [-WithVex | -WithoutVex] [-NoModifyPath]
Default: install Wave and install Vex when its latest release supports this platform.
-WithVex requires Vex; -WithoutVex installs Wave only.
WAVE_INSTALL_DIR overrides the dedicated installation directory (%LOCALAPPDATA%\Wave\bin).
The matching standard library also installs under HOME (or USERPROFILE) at .wave\lib\wave\std.
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
    $installDir = if ($env:WAVE_INSTALL_DIR) { [IO.Path]::GetFullPath($env:WAVE_INSTALL_DIR) } else { [IO.Path]::GetFullPath((Join-Path $env:LOCALAPPDATA 'Wave\bin')) }
    $stdDir = Get-LegacyStdRoot
    if ($installDir -match '[;\r\n]') { throw 'Use an installation directory without semicolons or newlines.' }
    $installPath = Resolve-InstallerPath $installDir
    if ($installDir.TrimEnd([char[]]'\/') -eq [IO.Path]::GetPathRoot($installDir).TrimEnd([char[]]'\/') -or
        $installPath.Equals((Resolve-InstallerPath $HOME), [StringComparison]::OrdinalIgnoreCase)) { throw 'Use a dedicated installation directory.' }
    Assert-SeparateInstallationPaths $installDir $stdDir
    $existingInstall = Get-Item -LiteralPath $installDir -Force -ErrorAction SilentlyContinue
    if ($null -ne $existingInstall) {
        if ($existingInstall.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'The installation directory must not be a link or junction.' }
        if (-not (Test-Path -LiteralPath (Join-Path $installDir 'wavec.exe') -PathType Leaf)) {
            throw "Refusing to replace a directory not managed by Wave: $installDir"
        }
    }
    Assert-ManagedStdDirectory $stdDir
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
    Write-Info "Compiler default standard library: $stdDir"
    $parent = Split-Path -Parent $installDir
    $stdParent = Split-Path -Parent $stdDir
    $lockPath = "$installDir.install-lock"
    $stdLockPath = "$stdDir.install-lock"
    $lock = $null; $stdLock = $null
    $temp = Join-Path $parent ('.wave-install-' + [Guid]::NewGuid().ToString('N'))
    $stdTemp = Join-Path $stdParent ('.wave-std-' + [Guid]::NewGuid().ToString('N'))
    $stage = Join-Path $temp 'stage'; $backup = Join-Path $temp 'previous'
    $stdStage = Join-Path $stdTemp 'stage'; $stdBackup = Join-Path $stdTemp 'previous'
    $activated = $false; $stdActivated = $false; $committed = $false
    $keepBackup = $false; $keepStdBackup = $false
    try {
        New-Item -ItemType Directory -Force -Path $parent, $stdParent | Out-Null
        # Different custom compiler directories still share the compiler's per-user std.
        $lock = [IO.File]::Open($lockPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
        $stdLock = [IO.File]::Open($stdLockPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
        Assert-SeparateInstallationPaths $installDir $stdDir
        Assert-ManagedStdDirectory $stdDir
        New-Item -ItemType Directory -Path $temp, $stdTemp | Out-Null
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
        # Stage beside the legacy root, so activation and restore stay on its volume.
        Copy-Item -LiteralPath (Join-Path $stage 'std') -Destination $stdStage -Recurse
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
        if (Test-Path -LiteralPath $stdDir) { Move-Item -LiteralPath $stdDir -Destination $stdBackup }
        $stdActivated = $true
        Move-Item -LiteralPath $stdStage -Destination $stdDir
        $activated = $true
        Move-Item -LiteralPath $stage -Destination $installDir
        Write-Info '[4/4] Checking compiler, default std and runtime'
        Test-Installation $installDir $target ($null -ne $vexAsset) $temp
        $committed = $true
        if (-not $NoModifyPath) {
            try { Add-UserPath $installDir }
            catch { Write-Warning "Wave is installed, but PATH could not be configured. Add $installDir to PATH manually." }
        }
        Write-Info "Installed Wave $($wave.tag_name)."
        if ($vexAsset) { Write-Info "Installed Vex $($vex.tag_name)." }
    } finally {
        $restoreErrors = @()
        try {
            if (-not $committed) {
                # Attempt both restores even when one destination is inaccessible.
                try {
                    if ($activated -and (Test-Path -LiteralPath $installDir)) { Remove-Item -LiteralPath $installDir -Recurse -Force }
                    if (Test-Path -LiteralPath $backup) { Move-Item -LiteralPath $backup -Destination $installDir }
                } catch {
                    $keepBackup = Test-Path -LiteralPath $backup
                    $restoreErrors += "Previous compiler retained at $backup. $($_.Exception.Message)"
                }
                try {
                    if ($stdActivated -and (Test-Path -LiteralPath $stdDir)) { Remove-Item -LiteralPath $stdDir -Recurse -Force }
                    if (Test-Path -LiteralPath $stdBackup) { Move-Item -LiteralPath $stdBackup -Destination $stdDir }
                } catch {
                    $keepStdBackup = Test-Path -LiteralPath $stdBackup
                    $restoreErrors += "Previous standard library retained at $stdBackup. $($_.Exception.Message)"
                }
            }
            if (-not $keepBackup -and (Test-Path -LiteralPath $temp)) { Remove-Item -LiteralPath $temp -Recurse -Force }
            if (-not $keepStdBackup -and (Test-Path -LiteralPath $stdTemp)) { Remove-Item -LiteralPath $stdTemp -Recurse -Force }
            if ($restoreErrors.Count) { throw "Restore failed. $($restoreErrors -join ' ')" }
        } finally {
            # A failed CreateNew never gives this invocation ownership of that lock.
            try {
                if ($null -ne $stdLock) {
                    $stdLock.Dispose()
                    Remove-Item -LiteralPath $stdLockPath -Force
                }
            } finally {
                if ($null -ne $lock) {
                    $lock.Dispose()
                    Remove-Item -LiteralPath $lockPath -Force
                }
            }
        }
    }

}
# Dot-sourcing exposes helpers without installing anything.
if ($MyInvocation.InvocationName -ne '.') {
    try { Install-Wave }
    catch { Write-Error $_; exit 1 }
}
