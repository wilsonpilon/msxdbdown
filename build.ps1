<#
.SYNOPSIS
    Build script for MSX DB Down.

.DESCRIPTION
    Compiles msxdbdown for Windows and/or Linux with Release or Debug profiles.
    Optionally runs the resulting native binary with forwarded arguments.

.PARAMETER Windows
    Build for Windows (GOOS=windows). Output: ./dist/windows/msxdbdown.exe

.PARAMETER Linux
    Build for Linux (GOOS=linux). Output: ./dist/linux/msxdbdown
    Note: cross-compilation requires a compatible C toolchain (Fyne uses CGo).

.PARAMETER All
    Build for every supported platform (Windows + Linux).

.PARAMETER Release
    Strip debug symbols and enable size optimisations (-ldflags "-s -w").
    This is the default profile when neither -Release nor -Debug is given.

.PARAMETER DebugBuild
    Disable inlining and optimisations for debugger use (-gcflags "all=-N -l").
    No symbol stripping.

.PARAMETER Run
    After a successful build, run the binary for the current platform.

.PARAMETER RunArgs
    Arguments forwarded to the binary when -Run is used.
    Example:  -RunArgs "--lang","en","--theme","dark"
              -RunArgs "version"

.PARAMETER Clean
    Delete the ./dist directory before building.

.PARAMETER Version
    Version string embedded in the binary (default: "dev").

.EXAMPLE
    .\build.ps1 -Windows -Release
    .\build.ps1 -Linux   -DebugBuild
    .\build.ps1 -All     -Release  -Version "1.0.0"
    .\build.ps1 -Windows -Release  -Run  -RunArgs "--lang","en","--theme","dark"
    .\build.ps1 -Windows -DebugBuild    -Run  -RunArgs "version"
    .\build.ps1 -Clean
#>

[CmdletBinding()]
param(
    [switch]    $Windows,
    [switch]    $Linux,
    [switch]    $All,
    [switch]    $Release,
    [switch]    $DebugBuild,
    [switch]    $Run,
    [string[]]  $RunArgs = @(),
    [switch]    $Clean,
    [string]    $Version = "dev"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ---------------------------------------------------------------------------
# Output helpers
# ---------------------------------------------------------------------------

function Write-Banner([string]$msg) {
    Write-Host ""
    Write-Host "  =====================================================" -ForegroundColor DarkCyan
    Write-Host "   $msg" -ForegroundColor Cyan
    Write-Host "  =====================================================" -ForegroundColor DarkCyan
}
function Write-Step([string]$msg)    { Write-Host "  >> $msg" -ForegroundColor Yellow }
function Write-Ok([string]$msg)      { Write-Host "  OK $msg" -ForegroundColor Green }
function Write-Warn([string]$msg)    { Write-Host "  !! $msg" -ForegroundColor DarkYellow }
function Write-Fail([string]$msg)    { Write-Host "  XX $msg" -ForegroundColor Red }
function Write-Divider               { Write-Host "  -----------------------------------------------------" -ForegroundColor DarkGray }

function Format-Elapsed([System.TimeSpan]$ts) {
    if ($ts.TotalSeconds -lt 60) { return ("{0:F1}s" -f $ts.TotalSeconds) }
    return ("{0}m {1:F0}s" -f [int]$ts.Minutes, $ts.Seconds)
}

# ---------------------------------------------------------------------------
# Resolve defaults
# ---------------------------------------------------------------------------

if (-not $Release -and -not $DebugBuild) { $Release = $true }

$nativeOS   = if ($env:OS -eq "Windows_NT") { "windows" } else { "linux" }
$onWindows  = ($nativeOS -eq "windows")
$onLinux    = ($nativeOS -eq "linux")

if ($All) { $Windows = $true; $Linux = $true }

if (-not $Windows -and -not $Linux) {
    if ($nativeOS -eq "windows") { $Windows = $true } else { $Linux = $true }
}

$profileLabel = if ($DebugBuild) { "debug" } else { "release" }
$now          = [System.DateTime]::UtcNow
$buildDate    = $now.ToString("ddMMyyyy")   # Military date: DDMMYYYY  e.g. 29052026
$buildTime    = $now.ToString("HHmm")       # Military time: HHMM      e.g. 1433
$unixTime     = [int64]($now - [datetime]"1970-01-01T00:00:00Z").TotalSeconds
$buildNumber  = "{0:x}" -f $unixTime        # Hex of UTC Unix timestamp e.g. 6a19e3d1

$ldBase  = "-X main.AppVersion=$Version -X main.BuildDate=$buildDate -X main.BuildTime=$buildTime -X main.BuildNumber=$buildNumber"
$ldflags = if ($Release)    { "-s -w $ldBase" } else { $ldBase }
$gcflags = if ($DebugBuild) { "all=-N -l" }     else { "" }

$distDir = Join-Path $PSScriptRoot "dist"

# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------

Write-Banner "MSX DB Down -- build script"
Write-Host "  Profile  : " -NoNewline; Write-Host $profileLabel.ToUpper()       -ForegroundColor Magenta
Write-Host "  Version  : " -NoNewline; Write-Host $Version                      -ForegroundColor White
Write-Host "  Date     : " -NoNewline; Write-Host "$buildDate $buildTime (UTC)" -ForegroundColor White
Write-Host "  Build#   : " -NoNewline; Write-Host $buildNumber                  -ForegroundColor Cyan
Write-Host ""

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------

if ($Clean) {
    Write-Step "Cleaning dist/ ..."
    if (Test-Path $distDir) {
        Remove-Item -Recurse -Force $distDir
        Write-Ok "Removed $distDir"
    } else {
        Write-Ok "dist/ does not exist -- nothing to clean"
    }
    if (-not $Windows -and -not $Linux) { Write-Host ""; exit 0 }
}

New-Item -ItemType Directory -Force -Path $distDir | Out-Null

# ---------------------------------------------------------------------------
# Build function
# ---------------------------------------------------------------------------

function Invoke-GoBuild([string]$goos, [string]$goarch = "amd64") {
    $label   = "$goos/$goarch"
    $outDir  = Join-Path $distDir $goos
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null

    $binName = if ($goos -eq "windows") { "msxdbdown.exe" } else { "msxdbdown" }
    $outPath = Join-Path $outDir $binName

    Write-Step "Building $label  -->  $outPath"

    $sw = [System.Diagnostics.Stopwatch]::StartNew()

    $env:GOOS        = $goos
    $env:GOARCH      = $goarch
    $env:CGO_ENABLED = "1"

    if ($goos -ne $nativeOS) {
        Write-Warn "Cross-compilation ($nativeOS -> $goos) requires a matching C toolchain."
        Write-Warn "If this fails, try: go install github.com/fyne-io/fyne-cross@latest"
        if ($goos -eq "linux"   -and $onWindows) { $env:CC = "x86_64-linux-gnu-gcc"   }
        if ($goos -eq "windows" -and $onLinux)   { $env:CC = "x86_64-w64-mingw32-gcc" }
    } else {
        if (Test-Path Env:\CC) { Remove-Item Env:\CC }
    }

    $args = [System.Collections.Generic.List[string]]::new()
    $args.Add("build")

    if ($gcflags -ne "") { $args.Add("-gcflags"); $args.Add($gcflags) }

    $finalLd = $ldflags
    # Release Windows: hide console window (comment out to keep CLI output)
    if ($goos -eq "windows" -and $Release) { $finalLd = "$ldflags -H windowsgui" }

    $args.Add("-ldflags"); $args.Add($finalLd)
    $args.Add("-o");       $args.Add($outPath)
    $args.Add(".")

    $goExe = (Get-Command go -ErrorAction Stop).Source
    & $goExe @args
    $code = $LASTEXITCODE

    $sw.Stop()
    foreach ($v in @("GOOS","GOARCH","CGO_ENABLED","CC")) { Remove-Item "Env:\$v" -ErrorAction SilentlyContinue }

    if ($code -ne 0) {
        Write-Fail "Build FAILED for $label (exit $code)"
        return $null
    }

    $bytes   = (Get-Item $outPath).Length
    $sizeStr = if ($bytes -ge 1MB) { ("{0:F1} MB" -f ($bytes / 1MB)) } else { ("{0:F0} KB" -f ($bytes / 1KB)) }
    Write-Ok "$label  [$sizeStr]  $(Format-Elapsed $sw.Elapsed)"
    return $outPath
}

# ---------------------------------------------------------------------------
# Run builds
# ---------------------------------------------------------------------------

$built = @{}

if ($Windows) { $p = Invoke-GoBuild "windows"; if ($p) { $built["windows"] = $p } }
if ($Linux)   { $p = Invoke-GoBuild "linux";   if ($p) { $built["linux"]   = $p } }

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

Write-Host ""
Write-Divider

if ($built.Count -eq 0) { Write-Fail "No binaries were produced."; exit 1 }

Write-Host "  Binaries:" -ForegroundColor Cyan
foreach ($kv in $built.GetEnumerator()) {
    Write-Host ("    {0,-10} {1}" -f $kv.Key, $kv.Value) -ForegroundColor White
}

# ---------------------------------------------------------------------------
# Run
# ---------------------------------------------------------------------------

if ($Run) {
    $bin = $built[$nativeOS]
    if (-not $bin) {
        Write-Host ""
        Write-Warn "-Run was requested but $nativeOS binary was not built."
        $flag = $nativeOS.Substring(0,1).ToUpper() + $nativeOS.Substring(1)
        Write-Warn "Add -$flag to include the native platform in this build."
    } else {
        Write-Host ""
        Write-Step "Running: $bin $RunArgs"
        Write-Divider
        & $bin @RunArgs
    }
}

Write-Host ""








