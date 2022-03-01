# Bhojpur Application - Foundation Framework

The `Bhojpur Application` is a primary framework applied within the [Bhojpur.NET Platform](https://github.com/bhojpur/platform) ecosystem to build web-scale, multi-cloud applications, and/or services for the distributed enterprises. We offer them in a __software-as-a-service__ (SaaS) delivery model. It leverages the [Bhojpur ORM](https://github.com/bhojpur/orm) engine to integrate with various relational databases.

## Client-side Access Framework

It features [Bhojpur IAM](https://github.com/bhojpur/iam) integration to support enterprise security management.

## Server-side Entitlement Framework

It features [Bhojpur CMS](https://github.com/bhojpur/cms) integration to support application-level entitlement management.

## Application Runtime Engine

The [Bhojpur Application](https://github.com/bhojpur/application) `runtime engine` is a portable, event-driven
integration middleware used for building distributed applications across __private-__ or __public-__ `Cloud`
and `Edge Computing` infrastructure. It manages execution of micro-services developed in a wide variety of
`programming languages`, `operating systems`, and `hosting environments` (e.g. web-browsers, Docker, Kubernetes
cluster). It securely manages IT systems by enforcing entitlement policy management for large enterprises.

```bash
$ appsvr init
```

The `runtime engine` operates in two different modes.

- `Standalone` mode is mostly used by software developers during their application development since it has a
relatively small footprint and easy to configure iteratively. However, it is applied in production too very
effectively in many [Bhojpur.NET Platform](https://github.com/bhojpur/platform) enabled systems.
- `Kubernetes` mode is used very often production due to integration with cloud service orchestrator (CSO),
distributed tracing systems.

### Usage

One of the __use-case__ is to support `wasm`-aware web applications. The `runtime engine` is embedded in the
hosting environment to allow secure execution of programs written in different programming languages.

## Command Line Interface

The [Bhojpur Application](https://github.com/bhojpur/application) `CLI` is a utility and client-side
command & control engine that manages [Bhojpur Application](https://github.com/bhojpur/application)
`runtime engine` instances in a distributed environment.

```bash
$ appctl init
```
