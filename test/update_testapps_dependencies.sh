#!/bin/bash

[ -z "$1" ] && echo "Usage: $0 [LAST_APP_COMMIT]" && exit 1

APP_CORE_COMMIT=$1

# Update Bhojpur Application dependencies for all E2E test apps
cd ./apps
appsroot=`pwd`
appsdirName='apps'
for appdir in * ; do
   if test -f "$appsroot/$appdir/go.mod"; then
      cd $appsroot/$appdir > /dev/null
      go get -u github.com/bhojpur/application@$APP_CORE_COMMIT
      go mod tidy
      echo "successfully updated Bhojpur Application dependency for $appdir"
   fi
done