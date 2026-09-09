#Requires -Version 5.1
<#
==============================================================================
 bgscan installer (Windows)
 https://github.com/MohsenBg/bgscan
------------------------------------------------------------------------------
 Installs bgscan via the native Rust installer (bgscan-installer). The
 installer is resolved from PATH or downloaded from its GitHub release, then
 delegates to `bgscan-installer install`, which resolves the latest (or a
 pinned) release and verifies its SHA-256 checksum.

 This script is fully standalone and safe to pipe:

   irm <raw-url> -UseBasicParsing | iex

 Usage:
   powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
   powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1 --version v2.10.0
==============================================================================
#>
$ErrorActionPreference = 'Stop'

$RepositoryOwner = 'MohsenBg'
$RepositoryName  = 'bgscan-installer'

$Version = 'latest'

# Manual argument parsing so the script works both as a file and piped via
# `iex`, matching install.sh semantics.
$i = 0
while ($i -lt $args.Count) {
    switch ($args[$i]) {
        { $_ -eq '--version' -or $_ -eq '-version' } {
            if ($i + 1 -ge $args.Count) { throw '--version requires a value' }
            $Version = $args[$i + 1]
            $i += 2
        }
        default {
            throw "Unknown argument: $($args[$i])"
        }
    }
}

function Get-BuilderFromPath {
    $cmd = Get-Command bgscan-installer -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    return $null
}

function Get-WindowsArch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }

    switch -Regex ($arch) {
        'AMD64' { return '64' }
        'ARM64' { return 'arm64' }
        'x86'   { return '32' }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Install-Builder {
    $asset = "bgscan-installer-windows-$(Get-WindowsArch).exe"
    $url = "https://github.com/$RepositoryOwner/$RepositoryName/releases/latest/download/$asset"

    $tmpDir = Join-Path $env:TEMP ("bgscan-installer-" + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tmpDir | Out-Null

    $builderPath = Join-Path $tmpDir 'bgscan-installer.exe'

    Write-Host "Downloading $RepositoryName ($asset) ..." -ForegroundColor DarkGray
    try {
        Invoke-WebRequest -Uri $url -OutFile $builderPath -UseBasicParsing
    }
    catch {
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
        throw "Failed to download $RepositoryName from $url"
    }

    return $builderPath
}

$builderPath = Get-BuilderFromPath
$downloadedTemp = $false
if (-not $builderPath) {
    Write-Host 'bgscan-installer not found on PATH; downloading it ...' -ForegroundColor DarkGray
    $builderPath = Install-Builder
    $downloadedTemp = $true
}

try {
    & $builderPath install --version $Version
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
finally {
    if ($downloadedTemp) {
        Remove-Item -Recurse -Force (Split-Path -Parent $builderPath) -ErrorAction SilentlyContinue
    }
}
