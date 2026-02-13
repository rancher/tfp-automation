param(
    [Parameter(Mandatory=$true)]
    [string]$Provider,

    [Parameter(Mandatory=$true)]
    [string]$Version
)

# Example:
# .\setup-provider.ps1 rancher2 v14.1.0-rc.2

$ErrorActionPreference = "Stop"

$VersionTag = $Version.Substring(1)

$Platform = "windows_amd64"

$PluginRoot = Join-Path $env:APPDATA "terraform.d\plugins\terraform.local\local"
$Dir = Join-Path $PluginRoot "$Provider\$VersionTag\$Platform"

Write-Host "Creating directory: $Dir"
New-Item -ItemType Directory -Force -Path $Dir | Out-Null

$ZipUrl = "https://github.com/rancher/terraform-provider-$Provider/releases/download/$Version/terraform-provider-$Provider`_$VersionTag`_$Platform.zip"
$ZipPath = "$env:TEMP\provider.zip"

Write-Host "Downloading provider from:"
Write-Host $ZipUrl

Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipPath

Write-Host "Extracting provider..."
Expand-Archive -Path $ZipPath -DestinationPath $Dir -Force

$Binary = Get-ChildItem -Path $Dir -Filter "terraform-provider-$Provider*_*.exe" -Recurse | Select-Object -First 1

if ($Binary) {
    $FinalPath = Join-Path $Dir "terraform-provider-$Provider.exe"

    Move-Item -Path $Binary.FullName -Destination $FinalPath -Force

    Write-Host ""
    Write-Host "Terraform provider $Provider $Version is ready to test!"
    Write-Host ""
    Write-Host "Use this in your Terraform config:"
    Write-Host ""
    Write-Host "terraform {"
    Write-Host "  required_providers {"
    Write-Host "    rancher2 = {"
    Write-Host "      source  = `"terraform.local/local/$Provider`""
    Write-Host "      version = `"$VersionTag`""
    Write-Host "    }"
    Write-Host "  }"
    Write-Host "}"
}
else {
    Write-Error "Provider binary not found after extraction!"
}