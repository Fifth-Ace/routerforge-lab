$ErrorActionPreference = "Stop"

$Repo = Split-Path -Parent $PSScriptRoot
Set-Location $Repo

$Go = Get-Command go.exe -ErrorAction SilentlyContinue
if (-not $Go) {
    $Fallback = "C:\Program Files\Go\bin\go.exe"
    if (-not (Test-Path $Fallback -PathType Leaf)) {
        throw "go.exe not found"
    }
    $GoExe = $Fallback
} else {
    $GoExe = $Go.Source
}

Write-Host "=== RouterForge Core v1 parity gate ==="
& $GoExe version

$ParityTests = @(
    "TestCoreParityHTTPBaseline",
    "TestAuthMiddlewareBlocksProtectedAPI",
    "TestAuthMiddlewareAllowsOnlyLoopbackModuleHealthWithoutSession",
    "TestAuthSessionAllowsProtectedAPI",
    "TestCatalogDetectsManagedAdminModule",
    "TestCatalogManagedMonitoringModules",
    "TestCatalogDetectsExternalIntegrations",
    "TestCatalogInstallPlansArePreviewOnly",
    "TestSafeCatalogPackageName",
    "TestExpandAssetTemplate",
    "TestParseChecksumList",
    "TestValidateCatalogActionCompletionRejectsVersionMismatch",
    "TestValidateCatalogActionCompletionAcceptsTargetVersion",
    "TestModuleTargetPathPreservesTrailingSlash",
    "TestModuleTargetPathUIRedirectRegression",
    "TestModuleUnavailableReportsInstalledDuringRestart",
    "TestModuleUnavailableReportsAbsentModule",
    "TestModuleUIUnavailableReturnsReconnectHTML",
    "TestModuleUIProxyStillServesLiveUnixSocket",
    "TestBundledRouterForgeRegistry",
    "TestRegistryRejectsRawShellLifecycle",
    "TestParseRouterForgeReleaseIndex",
    "TestParseRouterForgeReleaseIndexAcceptsRenamedRepository",
    "TestParseRouterForgeReleaseIndexRejectsForeignRepository",
    "TestCoreCanExposeIndependentUpdate"
)

$Pattern = "^(" + (($ParityTests | ForEach-Object { [regex]::Escape($_) }) -join "|") + ")$"

Write-Host ""
Write-Host "=== Contract parity vectors ==="
$ParityTests | ForEach-Object { Write-Host " - $_" }

Write-Host ""
Write-Host "=== Run parity vectors ==="

& $GoExe test ./components/core -run $Pattern -count=1 -v
if ($LASTEXITCODE -ne 0) {
    throw "Core v1 parity gate failed"
}

Write-Host ""
Write-Host "=== Run complete Core regression suite ==="

& $GoExe test ./components/core -count=1
if ($LASTEXITCODE -ne 0) {
    throw "Core regression suite failed"
}

Write-Host ""
Write-Host "CORE_V1_PARITY=PASS"
