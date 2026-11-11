#requires -Version 5.1
<#
==============================================================================
 bgscan dependency installer (Windows)
------------------------------------------------------------------------------
 Project:
   BgScan

 Purpose:
   Ensures bgscan-builder is installed and executes the dependency bootstrap
   process required for BgScan development.

 Workflow:
   1. Validate current directory is a BgScan project.
   2. Install bgscan-builder if missing (calls install-builder.ps1).
   3. Execute:

        .\bgscan-builder.exe setup-dev --project-dir <project-root>

 Notes:
   - Must be executed from the root of a BgScan project.
   - Requires go.mod to be present.
   - Intended for developers and CI environments (Windows runners).
==============================================================================
#>

$ErrorActionPreference = "Stop"

$ProjectRoot     = (Get-Location).Path
$BuilderBinary   = ".\bgscan-builder.exe"
$InstallScript   = ".\scripts\install-builder.ps1"
$ScriptName      = Split-Path -Leaf $PSCommandPath
$StartTime       = Get-Date
$Step            = 0

# ==============================================================================
# LOGGER
# ==============================================================================

function Get-Timestamp {
    Get-Date -Format "yyyy-MM-dd HH:mm:ss"
}

function Write-Step {
    param([string]$Message)
    $script:Step++
    Write-Host ""
    Write-Host ("-" * 64) -ForegroundColor Blue
    Write-Host ("[STEP {0}] {1}  {2}" -f $script:Step, (Get-Timestamp), $ScriptName) -ForegroundColor Blue
    Write-Host ("-" * 64) -ForegroundColor Blue
    Write-Host "  $Message"
}

function Write-Info {
    param([string]$Message)
    Write-Host ("  {0} [INFO]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Green
}

function Write-WarnLog {
    param([string]$Message)
    Write-Host ("  {0} [WARN]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host ("  {0} [ OK ]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host ""
    Write-Host ("-" * 64) -ForegroundColor Red
    Write-Host ("[FATAL] {0}  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Red
    Write-Host ("-" * 64) -ForegroundColor Red
    Write-Host ""
    exit 1
}

# ==============================================================================
# PROJECT VALIDATION
# ==============================================================================

function Validate-Project {
    Write-Step "Validating BgScan project structure"

    $goModPath = Join-Path $ProjectRoot "go.mod"
    Write-Info "Checking for go.mod in: $ProjectRoot"
    if (-not (Test-Path $goModPath)) {
        Write-Fail "go.mod not found in $ProjectRoot. Are you in the project root?"
    }
    Write-Success "go.mod located"

    $moduleLine = Select-String -Path $goModPath -Pattern '^module\s+(\S+)' | Select-Object -First 1
    if (-not $moduleLine) {
        Write-Fail "Unable to determine module name from go.mod"
    }
    $moduleName = $moduleLine.Matches[0].Groups[1].Value
    Write-Info "Module name resolved: $moduleName"

    if ($moduleName -notmatch "bgscan") {
        Write-Fail "Unsupported module '$moduleName'. This script must be run inside a BgScan project."
    }

    Write-Success "Project validated - module: $moduleName"
}

# ==============================================================================
# BUILDER INSTALLATION
# ==============================================================================

function Ensure-Builder {
    Write-Step "Checking for bgscan-builder binary"

    if (Test-Path $BuilderBinary) {
        Write-Success "bgscan-builder already present at $BuilderBinary"
        return
    }

    Write-WarnLog "bgscan-builder not found at $BuilderBinary"
    Write-Info "Looking for installer script: $InstallScript"

    if (-not (Test-Path $InstallScript)) {
        Write-Fail "Installer script not found at $InstallScript. Cannot proceed without it."
    }

    Write-Info "Running installer: $InstallScript"
    & $InstallScript

    if (-not (Test-Path $BuilderBinary)) {
        Write-Fail "Installer completed but $BuilderBinary still not found"
    }

    Write-Success "bgscan-builder installed successfully"
}

# ==============================================================================
# TARGET LIST TEMPLATES
# ==============================================================================

function Stage-TargetTemplates {
    Write-Step "Staging target list templates for bgscan-builder"

    $srcDir = Join-Path $ProjectRoot "targets\ips"
    $dstDir = Join-Path $ProjectRoot "ips"

    $templates = @(Get-ChildItem -Path $srcDir -Filter "*.default" -File -ErrorAction SilentlyContinue)
    if ($templates.Count -eq 0) {
        Write-Info "No .default templates found in $srcDir"
        return
    }

    New-Item -ItemType Directory -Force -Path $dstDir | Out-Null
    foreach ($template in $templates) {
        Copy-Item -Path $template.FullName -Destination (Join-Path $dstDir $template.Name) -Force
    }

    Write-Success "Staged $($templates.Count) template(s) in $dstDir"
}

# ==============================================================================
# DEPENDENCY SETUP
# ==============================================================================

function Setup-Dependencies {
    Write-Step "Preparing dependency bootstrap"

    Write-Info "Invoking: $BuilderBinary setup-dev --project-dir $ProjectRoot"
    & $BuilderBinary setup-dev --project-dir $ProjectRoot

    if ($LASTEXITCODE -ne 0) {
        Write-Fail "bgscan-builder exited with code $LASTEXITCODE"
    }

    Write-Success "Dependency bootstrap finished without errors"
}

# ==============================================================================
# MAIN
# ==============================================================================

function Main {
    Write-Step "Starting BgScan dependency installation"
    Write-Info "Project root: $ProjectRoot"

    Validate-Project
    Ensure-Builder
    Stage-TargetTemplates
    Setup-Dependencies

    $elapsed = [math]::Round(((Get-Date) - $StartTime).TotalSeconds, 1)

    Write-Host ""
    Write-Host ("=" * 64) -ForegroundColor Green
    Write-Host "  BgScan dependency installation completed successfully" -ForegroundColor Green
    Write-Host "  Total steps : $Step" -ForegroundColor Green
    Write-Host "  Elapsed time: ${elapsed}s" -ForegroundColor Green
    Write-Host "  Finished at : $(Get-Timestamp)" -ForegroundColor Green
    Write-Host ("=" * 64) -ForegroundColor Green
    Write-Host ""
}

Main
