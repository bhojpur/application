# Bhojpur Application - CLI Installer

It provides ability to install key Bhojpur Application CLI software utilities.

## Windows

### Get the latest stable version

```sh
powershell -Command "iwr -useb https://raw.githubusercontent.com/bhojpur/application/master/install/install.ps1 | iex"
```

### Get the specific version

```sh
powershell -Command "$script=iwr -useb https://raw.githubusercontent.com/bhojpur/application/master/install/install.ps1; $block=[ScriptBlock]::Create($script); invoke-command -ScriptBlock $block -ArgumentList <Version>"
```

## MacOS

### Get the latest stable version

```sh
curl -fsSL https://raw.githubusercontent.com/bhojpur/application/master/install/install.sh | /bin/bash
```

### Get the specific version

```sh
curl -fsSL https://raw.githubusercontent.com/bhojpur/application/master/install/install.sh | /bin/bash -s <Version>
```

## Linux

### Get the latest stable version

```sh
wget -q https://raw.githubusercontent.com/bhojpur/application/master/install/install.sh -O - | /bin/bash
```

### Get the specific version

```sh
wget -q https://raw.githubusercontent.com/bhojpur/application/master/install/install.sh -O - | /bin/bash -s <Version>
```

## For Users with Poor Network Conditions

You can download resources from a mirror instead of from GitHub.

### Windows

- Create a `CustomAssetFactory` function to define what the release asset url you want to use
- Set `AppReleaseJsonUrl` to the equivalent of the json representation of all releases at <https://api.github.com/repos/bhojpur/application/releases>
- You could use cdn.jsdelivr.net global CDN for your location to download install.ps1

### Get the latest stable version

For example, if you are in Chinese mainland, you could use:

- Gitee.com to get latest release.json
- cnpmjs.org hosted by Alibaba for assets
- cdn.jsdelivr.net global CDN for install.ps1

```powershell
function CustomAssetFactory {
    param (
        $release
    )
    [hashtable]$return = @{}
    $return.url = "https://github.com.cnpmjs.org/bhojpur/application/releases/download/$($release.tag_name)/appctl_windows_amd64.zip"
    $return.name = "appctl_windows_amd64.zip"
    $return
}
$params = @{
    CustomAssetFactory = ${function:CustomAssetFactory};
    AppReleaseJsonUrl    = "https://gitee.com/app-cn/app-bin-mirror/raw/main/application/releases.json";
}
$script=iwr -useb https://cdn.jsdelivr.net/gh/bhojpur/application/install/install.ps1;
$block=[ScriptBlock]::Create(".{$script} $(&{$args} @params)");
Invoke-Command -ScriptBlock $block
```

### Get the specific version

```powershell
function CustomAssetFactory {
    param (
        $release
    )
    [hashtable]$return = @{}
    $return.url = "https://github.com.cnpmjs.org/bhojpur/application/releases/download/$($release.tag_name)/appctl_windows_amd64.zip"
    $return.name = "appctl_windows_amd64.zip"
    $return
}
$params = @{
    CustomAssetFactory = ${function:CustomAssetFactory};
    AppReleaseJsonUrl    = "https://gitee.com/app-cn/app-bin-mirror/raw/main/application/releases.json";
    Version            = <Version>
}
$script=iwr -useb https://cdn.jsdelivr.net/gh/bhojpur/application/install/install.ps1;
$block=[ScriptBlock]::Create(".{$script} $(&{$args} @params)");
Invoke-Command -ScriptBlock $block
```
