param (
    [string]$Version,
    [string]$AppRoot = "c:\bhojpur\application",
    [string]$AppReleaseJsonUrl = "",
    [scriptblock]$CustomAssetFactory = $null
)

Write-Output ""
$ErrorActionPreference = 'stop'

#Escape space of AppRoot path
$AppRoot = $AppRoot -replace ' ', '` '

# Constants
$AppCliFileName = "appctl.exe"
$AppCliFilePath = "${AppRoot}\${AppCliFileName}"

# GitHub Org and repo hosting Bhojpur Application CLI
$GitHubOrg = "bhojpur"
$GitHubRepo = "application"

# Set Github request authentication for basic authentication.
if ($Env:GITHUB_USER) {
    $basicAuth = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($Env:GITHUB_USER + ":" + $Env:GITHUB_TOKEN));
    $githubHeader = @{"Authorization" = "Basic $basicAuth" }
}
else {
    $githubHeader = @{}
}

if ((Get-ExecutionPolicy) -gt 'RemoteSigned' -or (Get-ExecutionPolicy) -eq 'ByPass') {
    Write-Output "PowerShell requires an execution policy of 'RemoteSigned'."
    Write-Output "To make this change please run:"
    Write-Output "'Set-ExecutionPolicy RemoteSigned -scope CurrentUser'"
    break
}

# Change security protocol to support TLS 1.2 / 1.1 / 1.0 - old powershell uses TLS 1.0 as a default protocol
[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"

# Check if Bhojpur Application CLI is installed.
if (Test-Path $AppCliFilePath -PathType Leaf) {
    Write-Warning "Bhojpur Application is detected - $AppCliFilePath"
    Invoke-Expression "$AppCliFilePath --version"
    Write-Output "Reinstalling Bhojpur Application..."
}
else {
    Write-Output "Installing Bhojpur Application..."
}

# Create Bhojpur Application Directory
Write-Output "Creating $AppRoot directory"
New-Item -ErrorAction Ignore -Path $AppRoot -ItemType "directory"
if (!(Test-Path $AppRoot -PathType Container)) {
    Write-Warning "Please visit https://docs.bhojpur.net/getting-started/install-app-cli/ for instructions on how to install without admin rights."
    throw "Cannot create $AppRoot"
}

# Get the list of release from GitHub
$releaseJsonUrl = $AppReleaseJsonUrl
if (!$releaseJsonUrl) {
    $releaseJsonUrl = "https://api.github.com/repos/${GitHubOrg}/${GitHubRepo}/releases"
}

$releases = Invoke-RestMethod -Headers $githubHeader -Uri $releaseJsonUrl -Method Get
if ($releases.Count -eq 0) {
    throw "No releases from github.com/bhojpur/application repo"
}

# get latest or specified version info from releases
function GetVersionInfo {
    param (
        [string]$Version,
        $Releases
    )
    # Filter windows binary and download archive
    if (!$Version) {
        $release = $Releases | Where-Object { $_.tag_name -notlike "*rc*" } | Select-Object -First 1
    }
    else {
        $release = $Releases | Where-Object { $_.tag_name -eq "v$Version" } | Select-Object -First 1
    }

    return $release
}

# get info about windows asset from release
function GetWindowsAsset {
    param (
        $Release
    )
    if ($CustomAssetFactory) {
        Write-Output "CustomAssetFactory dectected, try to invoke it"
        return $CustomAssetFactory.Invoke($Release)
    }
    else {
        $windowsAsset = $Release | Select-Object -ExpandProperty assets | Where-Object { $_.name -Like "*windows_amd64.zip" }
        if (!$windowsAsset) {
            throw "Cannot find the windows Bhojpur Application CLI binary"
        }
        [hashtable]$return = @{}
        $return.url = $windowsAsset.url
        $return.name = $windowsAsset.name
        return $return
    }`
}

$release = GetVersionInfo -Version $Version -Releases $releases
if (!$release) {
    throw "Cannot find the specified Bhojpur Application CLI binary version"
}
$asset = GetWindowsAsset -Release $release
$zipFileUrl = $asset.url
$assetName = $asset.name

$zipFilePath = $AppRoot + "\" + $assetName
Write-Output "Downloading $zipFileUrl ..."

$githubHeader.Accept = "application/octet-stream"
$oldProgressPreference = $progressPreference;
$progressPreference = 'SilentlyContinue';
Invoke-WebRequest -Headers $githubHeader -Uri $zipFileUrl -OutFile $zipFilePath
$progressPreference = $oldProgressPreference;
if (!(Test-Path $zipFilePath -PathType Leaf)) {
    throw "Failed to download Bhojpur Application Cli binary - $zipFilePath"
}

# Extract Bhojpur Application CLI to $AppRoot
Write-Output "Extracting $zipFilePath..."
Microsoft.Powershell.Archive\Expand-Archive -Force -Path $zipFilePath -DestinationPath $AppRoot
if (!(Test-Path $AppCliFilePath -PathType Leaf)) {
    throw "Failed to download Bhojpur Application Cli archieve - $zipFilePath"
}

# Check the Bhojpur Application CLI version
Invoke-Expression "$AppCliFilePath --version"

# Clean up zipfile
Write-Output "Clean up $zipFilePath..."
Remove-Item $zipFilePath -Force

# Add AppRoot directory to User Path environment variable
Write-Output "Try to add $AppRoot to User Path Environment variable..."
$UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPathEnvironmentVar -like '*app*') {
    Write-Output "Skipping to add $AppRoot to User Path - $UserPathEnvironmentVar"
}
else {
    [System.Environment]::SetEnvironmentVariable("PATH", $UserPathEnvironmentVar + ";$AppRoot", "User")
    $UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Output "Added $AppRoot to User Path - $UserPathEnvironmentVar"
}

Write-Output "`r`nBhojpur Application CLI is installed successfully."
Write-Output "To get started with Bhojpur Application, please visit https://docs.bhojpur.net/getting-started/ ."
Write-Output "Ensure that Docker Desktop is set to Linux containers mode when you run Bhojpur Application in self-hosted mode."