$ErrorActionPreference = 'Stop'
$installer = Join-Path $PSScriptRoot '../public/install.ps1'
$tokens = $null; $errors = $null
$ast = [Management.Automation.Language.Parser]::ParseFile((Resolve-Path $installer), [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw $errors }
. $installer
function Assert($Condition, $Message) { if (-not $Condition) { throw $Message } }
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
$previousDir = $env:WAVE_INSTALL_DIR
try {
    $path = Join-Path $temp 'data'; [IO.File]::WriteAllText($path, 'fixture')
    $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
    Assert-Hash $path $hash.ToLowerInvariant()
    Must-Fail { Assert-Hash $path ('0'*64) }
    $release = @{assets=@(@{name='wave.zip';state='uploaded';digest="sha256:$hash"})}
    Assert ((Get-ReleaseAsset $release 'wave.zip').name -eq 'wave.zip') 'asset selection'
    Assert ($null -eq (Get-ReleaseAsset $release 'absent' $true)) 'optional asset'
    foreach ($bad in @($null,'sha256:bad',"sha512:$hash")) {
        $release.assets[0].digest=$bad
        Must-Fail { Get-ReleaseAsset $release 'wave.zip' }
    }
    $release.assets[0].digest="sha256:$hash"; $release.assets += $release.assets[0]
    Must-Fail { Get-ReleaseAsset $release 'wave.zip' }

    # Exercise the real preparation/swap/rollback against local ZIPs.
    # Only network, host detection, process execution and user PATH writes are mocked.
    function Get-NativeWindowsTarget { return $script:Target }
    function Save-Download($Uri, $Path) {
        if ($script:Failure -eq 'download') { throw 'download fixture failure' }
        Copy-Item -LiteralPath $script:Archives[$Uri.Split('/')[-1]] -Destination $Path
    }
    function Test-Installation($Directory, $Target, $InstallVex, $Work) {
        Assert (Test-Path -LiteralPath (Join-Path $Directory 'std/manifest.json')) 'bundled std absent'
        Assert ((Get-Content -LiteralPath (Join-Path $Directory 'wavec.exe')) -eq 'new compiler') 'new compiler absent'
        if ($script:Failure -eq 'smoke') { throw 'smoke fixture failure' }
    }
    function Add-UserPath($Directory) { throw 'fixture PATH failure' }
    function Get-LatestRelease($Repository) { return $script:Releases[$Repository] }
    function Get-PinnedWaveRelease { return $script:Releases['wavefnd/Wave'] }
    $script:Archives=@{}; $script:Releases=@{}
    foreach ($target in @('x86_64-pc-windows-msvc','aarch64-pc-windows-msvc')) {
        $script:Target=$target
        foreach ($repo in @('Wave','Vex')) {
            $tag = if ($repo -eq 'Wave') { 'v0.2.1-pre-beta' } else { 'v1.0.0' }
            $name="$($repo.ToLower())-$tag-$target"
            $folder=Join-Path $temp $name
            New-Item -ItemType Directory -Force -Path (Join-Path $folder 'llvm/bin'),(Join-Path $folder 'std') | Out-Null
            [IO.File]::WriteAllText((Join-Path $folder 'std/manifest.json'),'{}')
            [IO.File]::WriteAllText((Join-Path $folder 'wavec.exe'),'new compiler')
            if ($repo -eq 'Vex') { [IO.File]::WriteAllText((Join-Path $folder 'vex.exe'),'new vex') }
            $zip=Join-Path $temp "$name.zip"
            Compress-Archive -LiteralPath $folder -DestinationPath $zip
            $script:Archives["$name.zip"]=$zip
            $script:Releases["wavefnd/$repo"]=@{tag_name=$tag;assets=@(@{name="$name.zip";state='uploaded';digest=('sha256:'+(Get-FileHash -LiteralPath $zip -Algorithm SHA256).Hash)})}
        }
        $env:WAVE_INSTALL_DIR=Join-Path $temp "install $target/bin"
        foreach ($failure in @('download','digest','smoke','none')) {
            $script:Failure=$failure
            New-Item -ItemType Directory -Force -Path $env:WAVE_INSTALL_DIR | Out-Null
            [IO.File]::WriteAllText((Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe'),'previous compiler')
            $saved=$script:Releases['wavefnd/Wave'].assets[0].digest
            if ($failure -eq 'digest') { $script:Releases['wavefnd/Wave'].assets[0].digest='sha256:'+('0'*64) }
            if ($failure -eq 'none') {
                Install-Wave
                Assert ((Get-Content (Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe')) -eq 'new compiler') 'activation failed'
            } else {
                Must-Fail { Install-Wave }
                Assert ((Get-Content (Join-Path $env:WAVE_INSTALL_DIR 'wavec.exe')) -eq 'previous compiler') 'rollback failed'
            }
            Assert (-not (Test-Path -LiteralPath "$env:WAVE_INSTALL_DIR.install-lock")) 'lock leaked'
            $script:Releases['wavefnd/Wave'].assets[0].digest=$saved
        }
        $script:Releases['wavefnd/Vex'].assets=@()
        $WithVex=$true; Must-Fail { Install-Wave } '*Latest Vex has no package*'; $WithVex=$false
        $NoModifyPath=$true; Install-Wave
        Assert (-not (Test-Path -LiteralPath (Join-Path $env:WAVE_INSTALL_DIR 'vex.exe'))) 'Vex should be absent'
        foreach ($old in @('1.0.0','nightly')) {
            $Version=$old; Must-Fail { Install-Wave } '*manually*'; $Version=''
        }
    }
} finally {
    $env:WAVE_INSTALL_DIR=$previousDir
    Remove-Item -LiteralPath $temp -Recurse -Force
}
Write-Host 'Pinned Wave and latest Vex PowerShell tests passed'
