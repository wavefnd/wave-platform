$ErrorActionPreference = 'Stop'
$script = Join-Path $PSScriptRoot '../public/install.ps1'
$tokens = $null
$errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $script), [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw $errors }
# Load only helper definitions; do not run any installation entry point.
foreach ($function in $ast.FindAll({param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst]}, $false)) {
    Invoke-Expression $function.Extent.Text
}
function Fail($Message) { throw $Message }
foreach ($tag in @('nightly', 'vnightly', 'NIGHTLY', 'vNightly')) {
    foreach ($helper in @('Normalize-Version', 'Assert-Version')) {
        $rejected = $false
        try { & $helper $tag } catch {
            $rejected = $_.Exception.Message -like '*manual download*'
        }
        if (-not $rejected) { throw "Accepted $tag through $helper" }
    }
}
if ((Normalize-Version '0.2.1-pre-beta') -ne 'v0.2.1-pre-beta') { throw 'version normalization changed' }
Assert-Version 'v0.2.1-pre-beta'
function Invoke-RestMethod {
    param($Headers, $Uri)
    return ,@(@{tag_name='nightly'; prerelease=$true}, @{tag_name='v0.2.1-pre-beta'; prerelease=$true})
}
if ((Resolve-LatestVersion 'wavefnd/Wave') -ne 'v0.2.1-pre-beta') { throw 'versioned prerelease not selected' }
function Invoke-RestMethod {
    param($Headers, $Uri)
    if ($Uri.EndsWith('page=1')) { return ,@(@{tag_name='nightly'; prerelease=$true}) }
    return ,@(@{tag_name='v0.2.0-pre-beta'; prerelease=$true})
}
if ((Resolve-LatestVersion 'wavefnd/Wave') -ne 'v0.2.0-pre-beta') { throw 'pagination failed' }
$script:emptyRequests = 0
function Invoke-RestMethod {
    param($Headers, $Uri)
    $script:emptyRequests++
    if ($script:emptyRequests -gt 1) { throw 'unexpected extra page after empty array' }
    return ,@()
}
$rejected = $false
try { Resolve-LatestVersion 'wavefnd/Wave' } catch { $rejected = $true }
if (-not $rejected -or $script:emptyRequests -ne 1) { throw 'empty release list accepted or paginated' }
Write-Host 'installer Nightly policy tests passed'
