$ErrorActionPreference = 'Stop'
$installer = Join-Path $PSScriptRoot '../public/install.ps1'
$tokens = $null; $errors = $null
$ast = [Management.Automation.Language.Parser]::ParseFile((Resolve-Path $installer), [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw $errors }
. $installer
$script:Assertions = 0
function Assert($Condition, $Message) {
    $script:Assertions += 1
    if (-not $Condition) { throw $Message }
}
function Must-Fail([scriptblock]$Action, $Pattern='*') {
    $failed = $false
    try { & $Action } catch { $failed = $_.Exception.Message -like $Pattern }
    Assert $failed "Expected failure: $Pattern"
}
foreach ($architecture in @('AMD64','x64','x86_64')) {
    Assert ((Get-WindowsTarget $architecture) -eq 'x86_64-pc-windows-msvc') 'amd64 target'
}
Assert ((Get-WindowsTarget 'ARM64') -eq 'aarch64-pc-windows-msvc') 'arm64 target'
Must-Fail { Get-WindowsTarget 'x86' }
$script:Responses = @()
function Invoke-RestMethod { param($Headers, $TimeoutSec, $Uri); return ,$script:Responses }
$script:Responses = @(
    @{id=1;tag_name='v1.0.0';draft=$false;published_at='2026-01-01T00:00:00Z'},
    @{id=2;tag_name='nightly';draft=$false;published_at='2026-10-05T00:00:00Z'},
    @{id=3;tag_name='v2.0.0';draft=$true;published_at='2026-10-05T00:00:00Z'},
    @{id=4;tag_name='v1.1.0-alpha';draft=$false;prerelease=$true;published_at='2026-10-01T00:00:00Z'}
)
Assert ((Get-LatestRelease 'wavefnd/Wave').tag_name -eq 'v1.1.0-alpha') 'latest public version'
$script:Responses = @()
Must-Fail { Get-LatestRelease 'wavefnd/Wave' }
$script:Responses = @{tag_name='v0.2.1-pre-beta';draft=$false}
Assert ((Get-PinnedWaveRelease).tag_name -eq 'v0.2.1-pre-beta') 'pinned Wave release'
foreach ($wrong in @(@{tag_name='nightly';draft=$false}, @{tag_name='v0.2.1-pre-beta';draft=$true})) {
    $script:Responses = $wrong
    Must-Fail { Get-PinnedWaveRelease } '*does not match*'
}
$temp = Join-Path ([IO.Path]::GetTempPath()) ('wave installer tests ' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp | Out-Null
$previousEnvironment = @{}
foreach ($name in @('HOME','USERPROFILE','LOCALAPPDATA','WAVE_INSTALL_DIR','WAVE_VERSION','VEX_VERSION')) {
    $previousEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
try {
    $env:HOME = Join-Path $temp 'home'
    $env:USERPROFILE = Join-Path $temp 'profile'
    $env:LOCALAPPDATA = Join-Path $temp 'local app data'
    $env:WAVE_VERSION = $null
    $env:VEX_VERSION = $null
    $path = Join-Path $temp 'data'
    [IO.File]::WriteAllText($path, 'fixture')
    $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
    Assert-Hash $path $hash.ToLowerInvariant()
    Must-Fail { Assert-Hash $path ('0'*64) }
    $release = @{assets=@(@{name='wave.zip';state='uploaded';digest="sha256:$hash"})}
    Assert ((Get-ReleaseAsset $release 'wave.zip').name -eq 'wave.zip') 'asset selection'
    Assert ($null -eq (Get-ReleaseAsset $release 'absent' $true)) 'optional asset'
    foreach ($bad in @($null,'sha256:bad',"sha512:$hash")) {
        $release.assets[0].digest = $bad
        Must-Fail { Get-ReleaseAsset $release 'wave.zip' }
    }
    $release.assets[0].digest = "sha256:$hash"
    $release.assets += $release.assets[0]
    Must-Fail { Get-ReleaseAsset $release 'wave.zip' }

    # Use local archives and isolated profile paths. Preparation, path checks,
    # swaps, rollback, and the default-lookup smoke routine remain real.
    function Get-NativeWindowsTarget { return $script:Target }
    function Save-Download($Uri, $Path) {
        if ($script:Failure -eq 'download') { throw 'download fixture failure' }
        Copy-Item -LiteralPath $script:Archives[$Uri.Split('/')[-1]] -Destination $Path
    }
    function Invoke-InstalledTool($Executable, [string[]]$Arguments) {
        $script:ToolInvocations += ,@($Executable, $Arguments)
        if ($Arguments[0] -eq '--version') { return 'fixture version' }
        if ($Arguments[0] -eq 'print' -and $Arguments[1] -eq 'target-spec') {
            return ('{"triple":"' + $script:Target + '"}')
        }
        if ($Arguments[0] -eq 'print' -and $Arguments[1] -eq 'std-path') {
            if ($script:Failure -eq 'std-path') { return (Join-Path $temp 'wrong std') }
            return Get-LegacyStdRoot
        }
        Assert ($Arguments[0] -eq 'run' -and $Arguments.Count -eq 2) 'smoke must use default std without an override'
        Assert ((Get-Content -LiteralPath (Join-Path (Get-LegacyStdRoot) 'manifest.json')) -eq '{"fixture":"packaged std"}') 'default std does not match package'
        Assert ((Get-Content -LiteralPath (Join-Path (Split-Path -Parent $Executable) 'std/manifest.json')) -eq '{"fixture":"packaged std"}') 'bundled std absent'
        Assert ((Get-Content -LiteralPath $Executable) -eq 'new compiler') 'new compiler absent'
        if ($script:Failure -in @('smoke','restore-compiler','restore-std')) { throw 'smoke fixture failure' }
    }
    function Move-Item($LiteralPath, $Destination) {
        $leaf = Split-Path -Leaf $LiteralPath
        if ($leaf -eq 'stage') {
            if ($Destination -eq (Get-LegacyStdRoot)) {
                Assert ((Split-Path -Parent (Split-Path -Parent $LiteralPath)) -eq (Split-Path -Parent (Get-LegacyStdRoot))) 'std staged on the wrong filesystem'
                if ($script:Failure -eq 'activate-std') { throw 'std activation fixture failure' }
            } elseif ($Destination -eq $env:WAVE_INSTALL_DIR -and $script:Failure -eq 'activate-compiler') {
                throw 'compiler activation fixture failure'
            }
        }
        if ($leaf -eq 'previous') {
            if (($script:Failure -eq 'restore-std' -and $Destination -eq (Get-LegacyStdRoot)) -or
                ($script:Failure -eq 'restore-compiler' -and $Destination -eq $env:WAVE_INSTALL_DIR)) {
                $script:RetainedBackup = $LiteralPath
                throw 'restore fixture failure'
            }
        }
        Microsoft.PowerShell.Management\Move-Item -LiteralPath $LiteralPath -Destination $Destination
        if ($leaf -eq 'stage' -and
            (($Destination -eq (Get-LegacyStdRoot) -and $script:Failure -eq 'activate-std-after') -or
            ($Destination -eq $env:WAVE_INSTALL_DIR -and $script:Failure -eq 'activate-compiler-after'))) {
            throw 'post-move activation fixture failure'
        }
    }
    function Add-UserPath($Directory) { throw 'fixture PATH failure' }
    function Get-LatestRelease($Repository) { return $script:Releases[$Repository] }
    function Get-PinnedWaveRelease { return $script:Releases['wavefnd/Wave'] }
    function Set-TestLocations($Name) {
        $caseRoot = Join-Path $temp $Name
        $env:HOME = Join-Path $caseRoot 'preferred home'
        $env:USERPROFILE = Join-Path $caseRoot 'fallback profile'
        $env:LOCALAPPDATA = Join-Path $caseRoot 'local app data'
        $env:WAVE_INSTALL_DIR = Join-Path $caseRoot 'custom install/bin'
        $script:ToolInvocations = @()
        $script:RetainedBackup = $null
        return $caseRoot
    }
    function Seed-PreviousInstallation {
        New-Item -ItemType Directory -Force -Path $env:WAVE_INSTALL_DIR, (Get-LegacyStdRoot) | Out-Null
        [IO.File]::WriteAllText((Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe'), 'previous compiler')
        [IO.File]::WriteAllText((Join-Path $env:WAVE_INSTALL_DIR 'compiler-sentinel'), 'previous compiler sentinel')
        [IO.File]::WriteAllText((Join-Path (Get-LegacyStdRoot) 'manifest.json'), '{"fixture":"previous std"}')
        [IO.File]::WriteAllText((Join-Path (Get-LegacyStdRoot) 'std-sentinel'), 'previous std sentinel')
    }
    function Assert-PreviousCompiler {
        Assert ((Get-Content -LiteralPath (Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe')) -eq 'previous compiler') 'compiler rollback failed'
        Assert (Test-Path -LiteralPath (Join-Path $env:WAVE_INSTALL_DIR 'compiler-sentinel')) 'compiler sentinel lost'
    }
    function Assert-PreviousStd {
        Assert ((Get-Content -LiteralPath (Join-Path (Get-LegacyStdRoot) 'manifest.json')) -eq '{"fixture":"previous std"}') 'std rollback failed'
        Assert (Test-Path -LiteralPath (Join-Path (Get-LegacyStdRoot) 'std-sentinel')) 'std sentinel lost'
    }
    function Assert-NoLocks {
        Assert (-not (Test-Path -LiteralPath "$env:WAVE_INSTALL_DIR.install-lock")) 'compiler lock leaked'
        Assert (-not (Test-Path -LiteralPath "$(Get-LegacyStdRoot).install-lock")) 'std lock leaked'
    }
    $script:Archives = @{}; $script:Releases = @{}
    foreach ($target in @('x86_64-pc-windows-msvc','aarch64-pc-windows-msvc')) {
        $script:Target = $target
        foreach ($repo in @('Wave','Vex')) {
            $tag = if ($repo -eq 'Wave') { 'v0.2.1-pre-beta' } else { 'v1.0.0' }
            $name = "$($repo.ToLower())-$tag-$target"
            $folder = Join-Path $temp $name
            New-Item -ItemType Directory -Force -Path (Join-Path $folder 'llvm/bin'),(Join-Path $folder 'std') | Out-Null
            [IO.File]::WriteAllText((Join-Path $folder 'std/manifest.json'), '{"fixture":"packaged std"}')
            [IO.File]::WriteAllText((Join-Path $folder 'wavec.exe'), 'new compiler')
            if ($repo -eq 'Vex') { [IO.File]::WriteAllText((Join-Path $folder 'vex.exe'), 'new vex') }
            $zip = Join-Path $temp "$name.zip"
            Compress-Archive -LiteralPath $folder -DestinationPath $zip
            $script:Archives["$name.zip"] = $zip
            $script:Releases["wavefnd/$repo"] = @{tag_name=$tag;assets=@(@{name="$name.zip";state='uploaded';digest=('sha256:'+(Get-FileHash -LiteralPath $zip -Algorithm SHA256).Hash)})}
        }
        foreach ($failure in @('download','digest','smoke','std-path','activate-std','activate-compiler','activate-std-after','activate-compiler-after','restore-std','restore-compiler','none')) {
            $caseRoot = Set-TestLocations "$target-$failure"
            $script:Failure = $failure
            Seed-PreviousInstallation
            $saved = $script:Releases['wavefnd/Wave'].assets[0].digest
            if ($failure -eq 'digest') { $script:Releases['wavefnd/Wave'].assets[0].digest = 'sha256:'+('0'*64) }
            if ($failure -eq 'none') {
                Install-Wave
                Assert ((Get-Content -LiteralPath (Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe')) -eq 'new compiler') 'activation failed'
                Assert ((Get-Content -LiteralPath (Join-Path (Get-LegacyStdRoot) 'manifest.json')) -eq '{"fixture":"packaged std"}') 'std activation failed'
                Assert (-not (Test-Path -LiteralPath (Join-Path $env:USERPROFILE '.wave/lib/wave/std'))) 'USERPROFILE overrode nonempty HOME'
            } else {
                Must-Fail { Install-Wave }
                if ($failure -eq 'restore-compiler') {
                    Assert (Test-Path -LiteralPath (Join-Path $script:RetainedBackup 'compiler-sentinel')) 'compiler backup lost after restore failure'
                } else { Assert-PreviousCompiler }
                if ($failure -eq 'restore-std') {
                    Assert (Test-Path -LiteralPath (Join-Path $script:RetainedBackup 'std-sentinel')) 'std backup lost after restore failure'
                } else { Assert-PreviousStd }
            }
            Assert-NoLocks
            $script:Releases['wavefnd/Wave'].assets[0].digest = $saved
        }
        $null = Set-TestLocations "$target-fresh"
        $script:Failure = 'none'
        Install-Wave
        Assert (Test-Path -LiteralPath (Join-Path (Get-LegacyStdRoot) 'manifest.json')) 'fresh std absent'
        Assert-NoLocks
        $null = Set-TestLocations "$target-fresh-smoke-failure"
        $script:Failure = 'smoke'
        Must-Fail { Install-Wave }
        Assert (-not (Test-Path -LiteralPath $env:WAVE_INSTALL_DIR)) 'fresh compiler left after smoke failure'
        Assert (-not (Test-Path -LiteralPath (Get-LegacyStdRoot))) 'fresh std left after smoke failure'
        Assert-NoLocks
        $script:Failure = 'none'
        $null = Set-TestLocations "$target-profile-fallback"
        $env:HOME = $null
        $env:WAVE_INSTALL_DIR = $null
        Install-Wave
        Assert (Test-Path -LiteralPath (Join-Path $env:USERPROFILE '.wave/lib/wave/std/manifest.json')) 'USERPROFILE fallback absent'
        Assert (Test-Path -LiteralPath (Join-Path $env:LOCALAPPDATA 'Wave/bin/wavec.exe')) 'default compiler path absent'
        $script:Releases['wavefnd/Vex'].assets = @()
        $WithVex = $true
        Must-Fail { Install-Wave } '*Latest Vex has no package*'
        $WithVex = $false
        $NoModifyPath = $true
        Install-Wave
        Assert (-not (Test-Path -LiteralPath (Join-Path $env:LOCALAPPDATA 'Wave/bin/vex.exe'))) 'Vex should be absent'
        foreach ($old in @('1.0.0','nightly')) {
            $Version = $old
            Must-Fail { Install-Wave } '*manually*'
            $Version = ''
        }
    }

    $null = Set-TestLocations 'shared-std-lock'
    Seed-PreviousInstallation
    $stdLockPath = "$(Get-LegacyStdRoot).install-lock"
    [IO.File]::WriteAllText($stdLockPath, 'another installer')
    Must-Fail { Install-Wave }
    Assert ((Get-Content -LiteralPath $stdLockPath) -eq 'another installer') 'foreign std lock removed'
    Assert (-not (Test-Path -LiteralPath "$env:WAVE_INSTALL_DIR.install-lock")) 'own compiler lock leaked after std lock conflict'
    Assert-PreviousCompiler
    Assert-PreviousStd
    $env:WAVE_INSTALL_DIR = Join-Path $temp 'another custom bin'
    Must-Fail { Install-Wave }
    Assert ((Get-Content -LiteralPath $stdLockPath) -eq 'another installer') 'custom root bypassed shared std lock'
    Remove-Item -LiteralPath $stdLockPath
    $null = Set-TestLocations 'compiler-lock'
    Seed-PreviousInstallation
    [IO.File]::WriteAllText("$env:WAVE_INSTALL_DIR.install-lock", 'another compiler installer')
    Must-Fail { Install-Wave }
    Assert ((Get-Content -LiteralPath "$env:WAVE_INSTALL_DIR.install-lock") -eq 'another compiler installer') 'foreign compiler lock removed'
    Assert (-not (Test-Path -LiteralPath "$(Get-LegacyStdRoot).install-lock")) 'std lock acquired before compiler lock conflict'
    Assert-PreviousCompiler
    Assert-PreviousStd

    foreach ($kind in @('directory','file')) {
        $null = Set-TestLocations "unmanaged-std-$kind"
        $stdDir = Get-LegacyStdRoot
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $stdDir) | Out-Null
        if ($kind -eq 'directory') {
            New-Item -ItemType Directory -Path $stdDir | Out-Null
            [IO.File]::WriteAllText((Join-Path $stdDir 'sentinel'), 'unmanaged')
        } else { [IO.File]::WriteAllText($stdDir, 'unmanaged') }
        Must-Fail { Install-Wave } '*not managed by Wave*'
        Assert-NoLocks
    }
    $null = Set-TestLocations 'overlap'
    $stdDir = Get-LegacyStdRoot
    foreach ($unsafe in @($stdDir, (Join-Path $stdDir 'bin'), (Split-Path -Parent $stdDir), (Join-Path $stdDir '../std/bin'))) {
        $env:WAVE_INSTALL_DIR = $unsafe
        Must-Fail { Install-Wave } '*must not overlap*'
    }
    $caseRoot = Set-TestLocations 'link-overlap'
    $stdParent = Split-Path -Parent (Get-LegacyStdRoot)
    New-Item -ItemType Directory -Force -Path $stdParent | Out-Null
    $aliasPath = Join-Path $caseRoot 'alias'
    $linkType = if ([Environment]::OSVersion.Platform -eq [PlatformID]::Win32NT) { 'Junction' } else { 'SymbolicLink' }
    New-Item -ItemType $linkType -Path $aliasPath -Target $stdParent | Out-Null
    $env:WAVE_INSTALL_DIR = Join-Path $aliasPath 'std/bin'
    Must-Fail { Install-Wave } '*must not overlap*'
    $caseRoot = Set-TestLocations 'compiler-link'
    $targetDir = Join-Path $caseRoot 'compiler link target'
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $env:WAVE_INSTALL_DIR), $targetDir | Out-Null
    New-Item -ItemType $linkType -Path $env:WAVE_INSTALL_DIR -Target $targetDir | Out-Null
    Must-Fail { Install-Wave } '*link or junction*'
    Remove-Item -LiteralPath $targetDir
    Must-Fail { Install-Wave } '*link or junction*'
    $caseRoot = Set-TestLocations 'std-link'
    $stdDir = Get-LegacyStdRoot
    $targetDir = Join-Path $caseRoot 'separate target'
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $stdDir), $targetDir | Out-Null
    [IO.File]::WriteAllText((Join-Path $targetDir 'manifest.json'), '{}')
    [IO.File]::WriteAllText((Join-Path $targetDir 'sentinel'), 'target preserved')
    New-Item -ItemType $linkType -Path $stdDir -Target $targetDir | Out-Null
    Must-Fail { Install-Wave } '*link or junction*'
    Assert ((Get-Content -LiteralPath (Join-Path $targetDir 'sentinel')) -eq 'target preserved') 'std link target modified'
    $env:HOME = 'relative-home'
    Must-Fail { Install-Wave } '*must be an absolute*'
    $env:HOME = $null
    $env:USERPROFILE = 'relative-profile'
    Must-Fail { Install-Wave } '*must be an absolute*'
    if ([Environment]::OSVersion.Platform -eq [PlatformID]::Win32NT) {
        foreach ($relativeProfile in @('C:profile', 'C:', '\profile')) {
            $env:USERPROFILE = $relativeProfile
            Must-Fail { Get-LegacyStdRoot } '*must be an absolute*'
        }
        foreach ($devicePath in @('\\?\C:\Wave', '\\.\C:\Wave')) {
            Must-Fail { Resolve-InstallerPath $devicePath } '*Windows device paths*'
        }
        Assert ((Resolve-InstallerPath 'C:\') -eq 'C:\') 'drive root was made drive-relative'
    }
    $env:HOME = $null
    $env:USERPROFILE = $null
    Must-Fail { Get-LegacyStdRoot } '*HOME or USERPROFILE*'
} finally {
    foreach ($name in $previousEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $previousEnvironment[$name], 'Process')
    }
    Remove-Item -LiteralPath $temp -Recurse -Force
}
Write-Host "Pinned Wave and latest Vex PowerShell tests passed ($script:Assertions assertions)"
