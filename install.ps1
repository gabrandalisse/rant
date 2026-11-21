$ErrorActionPreference = "Stop"

$Repo = "gabrandalisse/rant"
$Binary = "rant.exe"
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    throw "Arquitectura de 32 bits no soportada."
}

$Url = "https://github.com/$Repo/releases/latest/download/rant_windows_$Arch.zip"
$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    $ZipPath = Join-Path $TmpDir "rant.zip"
    Write-Host "Descargando rant para windows/$Arch..."
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath
    Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\rant"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path (Join-Path $TmpDir $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    }

    $ProfileDir = Split-Path -Parent $PROFILE
    if (!(Test-Path $ProfileDir)) {
        New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
    }
    if (!(Test-Path $PROFILE)) {
        New-Item -ItemType File -Path $PROFILE -Force | Out-Null
    }

    $HookLine = 'Set-PSReadLineOption -AddToHistoryHandler { param($line); return -not ($line -match ''^\s*rant\b'') }'
    $CurrentProfile = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue
    if ($CurrentProfile -notlike "*Set-PSReadLineOption*rant*") {
        Add-Content -Path $PROFILE -Value "`n# Evitar guardar comandos rant en el historial`n$HookLine"
        Write-Host "Protección de historial agregada en $PROFILE"
    }

    Write-Host "Instalado en $InstallDir\$Binary. Reiniciá la terminal para aplicar los cambios."
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}