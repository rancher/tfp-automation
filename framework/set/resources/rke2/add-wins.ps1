param (
    [string]$K8S_VERSION,
    [string]$RKE2_SERVER_IP,
    [string]$RKE2_TOKEN
)

powershell.exe -Command "Start-Process PowerShell -Verb RunAs"
Enable-WindowsOptionalFeature -Online -FeatureName containers -All
Invoke-WebRequest -Uri https://raw.githubusercontent.com/rancher/rke2/master/install.ps1 -Outfile install.ps1

New-Item -Type Directory -Path C:\etc\rancher\rke2 -Force
New-Item -Type File -Path C:\etc\rancher\rke2\config.yaml -Force

$RKE2_SERVER_URL = "https://" + $RKE2_SERVER_IP + ":9345"

cmd.exe /c "(echo server: $RKE2_SERVER_URL && echo token: $RKE2_TOKEN && echo node-name: tfp-wins && echo tls-san: && echo   - $RKE2_SERVER_IP) > C:\etc\rancher\rke2\config.yaml"

$env:PATH+=";c:\var\lib\rancher\rke2\bin;c:\usr\local\bin"

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + ";C:\var\lib\rancher\rke2\bin;c:\usr\local\bin",
    [EnvironmentVariableTarget]::Machine)

.\install.ps1 -Version $K8S_VERSION
rke2.exe agent service --add
Start-Service rke2