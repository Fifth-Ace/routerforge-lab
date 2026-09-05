param(
    [Parameter(Mandatory=$true)]
    [string]$GoExe,

    [string]$OutputDir = ".rf-phase2-size"
)

$ErrorActionPreference = "Stop"

$Repo = Split-Path -Parent $PSScriptRoot
Set-Location $Repo

$Version = (& $GoExe version).Trim()
if ($Version -notmatch "go1\.21\.13") {
    throw "expected Go 1.21.13, got: $Version"
}

$Out = Join-Path $Repo $OutputDir
if (Test-Path $Out) {
    Remove-Item $Out -Recurse -Force
}
New-Item -ItemType Directory -Path $Out | Out-Null

$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"

$Variants = @(
    @{ Name="noembed-default"; Tags=""; Flags=@() },
    @{ Name="noembed-stripped"; Tags=""; Flags=@("-trimpath", "-ldflags=-s -w -buildid=") },
    @{ Name="embed-default"; Tags="embed_frontend"; Flags=@() },
    @{ Name="embed-stripped"; Tags="embed_frontend"; Flags=@("-trimpath", "-ldflags=-s -w -buildid=") }
)

$Rows = @()

foreach ($V in $Variants) {
    $Bin = Join-Path $Out ("routerforge-" + $V.Name)

    $Args = @("build", "-o", $Bin)
    if ($V.Tags) {
        $Args += @("-tags", $V.Tags)
    }
    $Args += $V.Flags
    $Args += "./components/core"

    Write-Host ("BUILD " + $V.Name)
    & $GoExe @Args
    if ($LASTEXITCODE -ne 0) {
        throw ("build failed: " + $V.Name)
    }

    $Item = Get-Item $Bin
    $Hash = (Get-FileHash $Bin -Algorithm SHA256).Hash.ToLowerInvariant()

    $Rows += [pscustomobject]@{
        Variant = $V.Name
        Bytes   = [int64]$Item.Length
        MiB     = [math]::Round($Item.Length / 1MB, 3)
        SHA256  = $Hash
        Path    = $Bin
    }

    (& $GoExe version -m $Bin) |
        Set-Content -Encoding UTF8 (Join-Path $Out ($V.Name + "-buildinfo.txt"))
}

$Default = Join-Path $Out "routerforge-noembed-default"

& $GoExe tool nm -size -sort size $Default 2>$null |
    Select-Object -Last 150 |
    Set-Content -Encoding UTF8 (Join-Path $Out "noembed-default-top-symbols.txt")

$Deps = & $GoExe list -deps ./components/core
$Deps | Set-Content -Encoding UTF8 (Join-Path $Out "core-deps.txt")

$Rows | ConvertTo-Json -Depth 3 |
    Set-Content -Encoding UTF8 (Join-Path $Out "sizes.json")

$Rows | Format-Table -AutoSize

$NoEmbedDefault  = ($Rows | Where-Object Variant -eq "noembed-default").Bytes
$NoEmbedStripped = ($Rows | Where-Object Variant -eq "noembed-stripped").Bytes
$EmbedDefault    = ($Rows | Where-Object Variant -eq "embed-default").Bytes
$EmbedStripped   = ($Rows | Where-Object Variant -eq "embed-stripped").Bytes

"NOEMBED_STRIP_SAVING_BYTES=" + ($NoEmbedDefault - $NoEmbedStripped)
"EMBED_STRIP_SAVING_BYTES=" + ($EmbedDefault - $EmbedStripped)
"FRONTEND_EMBED_DELTA_DEFAULT_BYTES=" + ($EmbedDefault - $NoEmbedDefault)
"FRONTEND_EMBED_DELTA_STRIPPED_BYTES=" + ($EmbedStripped - $NoEmbedStripped)
