# Bhojpur Application - Distributed Framework

It is a common interface of [Bhojpur.NET Platform](https://github.com/bhojpur/platform)
distributed applications or services.

| packages  | description                                                            |
|-----------|------------------------------------------------------------------------|
| common    | common protos that are imported by multiple packages                   |
| internals | gRPC and protobuf definitions, which is used for appication internal   |
| runtime   | application and callback services and its associated protobuf messages |
| operator  | application operator gRPC service                                      |
| placement | applicaion placement service                                           |
| sentry    | application sentry for CA service                                      |

## Client-side ProtoBuf Generation

1. Install `protoc` version: [v3.14.0](https://github.com/protocolbuffers/protobuf/releases/tag/v3.14.0)

2. Install `protoc-gen-go` and `protoc-gen-go-grpc`

```bash
make init-proto
```

3. Generate gRPC proto clients

```bash
make gen-proto
```

## Update E2E Test Apps

Whenever there are breaking changes in the `.proto` files, we need to update the E2E test apps
to use the correct version of [Bhojpur Application](https://github.com/bhojpur/application)
dependencies. It could be done by navigating to the tests folder and running the commands.

```bash
# Use the last commit of Bhojpur Application.
./update_testapps_dependencies.sh be08e5520173beb93e5d5f047dbde405e78db658
```

**Note**: On Windows, use the mingw tools to execute the bash script

Check in all the `go.mod` files for the test apps that have now been modified to point to the
latest [Bhojpur Application](https://github.com/bhojpur/application) version.
