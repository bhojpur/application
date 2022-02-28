#!/usr/bin/env bash

# Bhojpur Application CLI location
: ${APP_INSTALL_DIR:="/usr/local/bin"}

# sudo is required to copy binary to APP_INSTALL_DIR for linux
: ${USE_SUDO:="false"}

# Http request CLI
APP_HTTP_REQUEST_CLI=curl

# GitHub Organization and repo name to download release
GITHUB_ORG=bhojpur
GITHUB_REPO=application

# Bhojpur Application CLI filename
APP_CLI_FILENAME=appctl

APP_CLI_FILE="${APP_INSTALL_DIR}/${APP_CLI_FILENAME}"

getSystemInfo() {
    ARCH=$(uname -m)
    case $ARCH in
        armv7*) ARCH="arm";;
        aarch64) ARCH="arm64";;
        x86_64) ARCH="amd64";;
    esac

    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    # Most linux distro needs root permission to copy the file to /usr/local/bin
    if [[ "$OS" == "linux" || "$OS" == "darwin" ]] && [ "$APP_INSTALL_DIR" == "/usr/local/bin" ]; then
        USE_SUDO="true"
    fi
}

verifySupported() {
    releaseTag=$1
    local supported=(darwin-amd64 linux-amd64 linux-arm linux-arm64)
    local current_osarch="${OS}-${ARCH}"

    for osarch in "${supported[@]}"; do
        if [ "$osarch" == "$current_osarch" ]; then
            echo "Your system is ${OS}_${ARCH}"
            return
        fi
    done

    if [ "$current_osarch" == "darwin-arm64" ]; then
        if isReleaseAvailable $releaseTag; then
            return
        else
            echo "The darwin_arm64 arch has no native binary for this version of Bhojpur Application, however you can use the amd64 version so long as you have rosetta installed"
            echo "Use 'softwareupdate --install-rosetta' to install rosetta if you don't already have it"
            ARCH="amd64"
            return
        fi
    fi

    echo "No prebuilt binary for ${current_osarch}"
    exit 1
}

runAsRoot() {
    local CMD="$*"

    if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
        CMD="sudo $CMD"
    fi

    $CMD || {
        echo "Please visit https://docs.bhojpur.net/getting-started/install-app-cli/ for instructions on how to install without sudo."
        exit 1
    }
}

checkHttpRequestCLI() {
    if type "curl" > /dev/null; then
        APP_HTTP_REQUEST_CLI=curl
    elif type "wget" > /dev/null; then
        APP_HTTP_REQUEST_CLI=wget
    else
        echo "Either curl or wget is required"
        exit 1
    fi
}

checkExistingApp() {
    if [ -f "$APP_CLI_FILE" ]; then
        echo -e "\nBhojpur Application CLI is detected:"
        $APP_CLI_FILE version
        echo -e "Reinstalling Bhojpur Application CLI - ${APP_CLI_FILE}...\n"
    else
        echo -e "Installing Bhojpur Application CLI...\n"
    fi
}

getLatestRelease() {
    local appReleaseUrl="https://api.github.com/repos/${GITHUB_ORG}/${GITHUB_REPO}/releases"
    local latest_release=""

    if [ "$APP_HTTP_REQUEST_CLI" == "curl" ]; then
        latest_release=$(curl -s $appReleaseUrl | grep \"tag_name\" | grep -v rc | awk 'NR==1{print $2}' |  sed -n 's/\"\(.*\)\",/\1/p')
    else
        latest_release=$(wget -q --header="Accept: application/json" -O - $appReleaseUrl | grep \"tag_name\" | grep -v rc | awk 'NR==1{print $2}' |  sed -n 's/\"\(.*\)\",/\1/p')
    fi

    ret_val=$latest_release
}

downloadFile() {
    LATEST_RELEASE_TAG=$1

    APP_CLI_ARTIFACT="${APP_CLI_FILENAME}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_BASE="https://github.com/${GITHUB_ORG}/${GITHUB_REPO}/releases/download"
    DOWNLOAD_URL="${DOWNLOAD_BASE}/${LATEST_RELEASE_TAG}/${APP_CLI_ARTIFACT}"

    # Create the temp directory
    APP_TMP_ROOT=$(mktemp -dt app-install-XXXXXX)
    ARTIFACT_TMP_FILE="$APP_TMP_ROOT/$APP_CLI_ARTIFACT"

    echo "Downloading $DOWNLOAD_URL ..."
    if [ "$APP_HTTP_REQUEST_CLI" == "curl" ]; then
        curl -SsL "$DOWNLOAD_URL" -o "$ARTIFACT_TMP_FILE"
    else
        wget -q -O "$ARTIFACT_TMP_FILE" "$DOWNLOAD_URL"
    fi

    if [ ! -f "$ARTIFACT_TMP_FILE" ]; then
        echo "failed to download $DOWNLOAD_URL ..."
        exit 1
    fi
}

isReleaseAvailable() {
    LATEST_RELEASE_TAG=$1

    APP_CLI_ARTIFACT="${APP_CLI_FILENAME}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_BASE="https://github.com/${GITHUB_ORG}/${GITHUB_REPO}/releases/download"
    DOWNLOAD_URL="${DOWNLOAD_BASE}/${LATEST_RELEASE_TAG}/${APP_CLI_ARTIFACT}"

    if [ "$APP_HTTP_REQUEST_CLI" == "curl" ]; then
        httpstatus=$(curl -sSLI -o /dev/null -w "%{http_code}" "$DOWNLOAD_URL")
        if [ "$httpstatus" == "200" ]; then
            return 0
        fi
    else
        wget -q --spider "$DOWNLOAD_URL"
        exitstatus=$?
        if [ $exitstatus -eq 0 ]; then
            return 0
        fi
    fi
    return 1
}

installFile() {
    tar xf "$ARTIFACT_TMP_FILE" -C "$APP_TMP_ROOT"
    local tmp_root_app_cli="$APP_TMP_ROOT/$APP_CLI_FILENAME"

    if [ ! -f "$tmp_root_app_cli" ]; then
        echo "Failed to unpack Bhojpur Application CLI executable."
        exit 1
    fi

    if [ -f "$APP_CLI_FILE" ]; then
        runAsRoot rm "$APP_CLI_FILE"
    fi
    chmod o+x $tmp_root_app_cli
    runAsRoot cp "$tmp_root_app_cli" "$APP_INSTALL_DIR"

    if [ -f "$APP_CLI_FILE" ]; then
        echo "$APP_CLI_FILENAME installed into $APP_INSTALL_DIR successfully."

        $APP_CLI_FILE --version
    else 
        echo "Failed to install $APP_CLI_FILENAME"
        exit 1
    fi
}

fail_trap() {
    result=$?
    if [ "$result" != "0" ]; then
        echo "Failed to install Bhojpur Application CLI"
        echo "For support, go to https://desk.bhojpur-consulting.com"
    fi
    cleanup
    exit $result
}

cleanup() {
    if [[ -d "${APP_TMP_ROOT:-}" ]]; then
        rm -rf "$APP_TMP_ROOT"
    fi
}

installCompleted() {
    echo -e "\nTo get started with Bhojpur Application, please visit https://docs.bhojpur.net/getting-started/"
}

# -----------------------------------------------------------------------------
# main
# -----------------------------------------------------------------------------
trap "fail_trap" EXIT

getSystemInfo
checkHttpRequestCLI

if [ -z "$1" ]; then
    echo "Getting the latest Bhojpur Application CLI..."
    getLatestRelease
else
    ret_val=v$1
fi

verifySupported $ret_val
checkExistingApp

echo "Installing $ret_val Bhojpur Application CLI..."

downloadFile $ret_val
installFile
cleanup

installCompleted