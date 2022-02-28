#!/bin/bash

set +x

VERSION=3.10.0

# TODO: Consider using a Docker image for this?

# Generate proto buffers
# First arg is name of language (e.g. 'javascript')
# Second arg is name of tool (e.g. 'js')
# Third arg is path within the language directory to write to
# Fourth arg is args for the generator
# Remaining args are appended to the command.
generate() {
    language=${1}
    tool=${2}
    path=${3}
    args=${4}

    mkdir -p ${top_root}/sdk/${language}/${path}
    
    for proto_file in "appclient/appclient.proto" "v1/runtime/app.proto"; do
        echo "Generating ${language} for ${proto_file}"
        ${root}/bin/protoc --proto_path ${top_root}/pkg/api/ \
            --${tool}_out=${args}:${top_root}/sdk/${language}/${path} \
            ${top_root}/pkg/api/${proto_file} ${@:5}
    done
}

# Setup the directories
root=$(dirname "${BASH_SOURCE[0]}")
top_root=${root}/../..

# Detect OS
OS=""
full_os=""
if [[ "$OSTYPE" == "linux-gnu" ]]; then
        OS="linux"
        full_os="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="osx"
        full_os="macosx"
fi

file="protoc-${VERSION}-${OS}-x86_64.zip"

# Download and install tools.
wget "https://github.com/protocolbuffers/protobuf/releases/download/v${VERSION}/${file}" \
  -O ${root}/${file}

# Download Java gRPC plugin
java_grpc_plugin_file="protoc-gen-grpc-java-1.24.0-${OS}-x86_64.exe"
java_grpc_plugin_path=${root}/${java_grpc_plugin_file}

wget "https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/1.24.0/${java_grpc_plugin_file}" \
  -O ${java_grpc_plugin_path}

chmod +x ${java_grpc_plugin_path}
unzip ${root}/${file} -d ${root}

# find grpc_tools_node_protoc_plugin location
PROTOC_PLUGIN=$(which grpc_tools_node_protoc_plugin)
ts_grpc_plugin_file=$(which protoc-gen-ts)
dotnet_grpc_plugin_file="${HOME}/.nuget/packages/grpc.tools/2.24.0/tools/${full_os}_x64/grpc_csharp_plugin"

language=${1:-"all"}

if [[ "${language}" = "all" ]] || [[ "${language}" = "js" ]]; then
  # generate javascript
  mkdir -p ${top_root}/sdk/js
  generate js js src 'import_style=commonjs' \
    --plugin=protoc-gen-grpc=${PROTOC_PLUGIN} \
    --plugin=protoc-gen-ts=${ts_grpc_plugin_file} \
    --ts_out="service=grpc-node:${top_root}/sdk/js/src" \
    --grpc_out=${top_root}/sdk/js/src
fi

if [[ "$language" = "all" ]] || [[ "$language" = "java" ]]; then
  # generate java
  mkdir -p ${top_root}/sdk/java
  generate java java src/main/java '' \
    --plugin=protoc-gen-grpc-java=${java_grpc_plugin_path} \
    --grpc-java_out=${top_root}/sdk/java/src/main/java
fi

if [[ "$language" = "all" ]] || [[ "$language" = "python" ]]; then
  # generate python
  mkdir -p ${top_root}/sdk/python
  python3 -m grpc.tools.protoc -I${top_root}/pkg/api \
   --python_out=${top_root}/sdk/python \
   --grpc_python_out=${top_root}/sdk/python \
   pkg/api/v1/runtime/app.proto \
   appclient/appclient.proto
fi

if [[ "$language" = "all" ]] || [[ "$language" = "go" ]]; then
  # generate golang
  generate go go . `` --plugin=grpc
fi

if [[ "$language" = "all" ]] || [[ "$language" = "dotnet" ]]; then
  # generate dotnet
  mkdir -p ${top_root}/sdk/dotnet
  # dotnet generates their own via dotnet build...
  generate dotnet csharp src/app.Client.Grpc '' \
    --plugin=protoc-gen-grpc=${dotnet_grpc_plugin_file} \
    --grpc_out=${top_root}/sdk/dotnet/src/app.Client.Grpc
fi

# cleanup
rm -r ${root}/include ${root}/bin ${root}/${file} ${root}/readme.txt ${java_grpc_plugin_path}