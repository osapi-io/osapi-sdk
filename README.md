[![release](https://img.shields.io/github/release/osapi-io/osapi-sdk.svg?style=for-the-badge)](https://github.com/osapi-io/osapi-sdk/releases/latest)
[![codecov](https://img.shields.io/codecov/c/github/osapi-io/osapi-sdk?style=for-the-badge)](https://codecov.io/gh/osapi-io/osapi-sdk)
[![go report card](https://goreportcard.com/badge/github.com/osapi-io/osapi-sdk?style=for-the-badge)](https://goreportcard.com/report/github.com/osapi-io/osapi-sdk)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/osapi-io/osapi-sdk/go.yml?style=for-the-badge)](https://github.com/osapi-io/osapi-sdk/actions/workflows/go.yml)
[![powered by](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![built with just](https://img.shields.io/badge/Built_with-Just-black?style=for-the-badge&logo=just&logoColor=white)](https://just.systems)
![gitHub commit activity](https://img.shields.io/github/commit-activity/m/osapi-io/osapi-sdk?style=for-the-badge)

# OSAPI SDK

Go SDK for [OSAPI][] — client library and orchestration primitives.

## Usage

https://github.com/osapi-io/osapi-sdk/blob/fe66750e11c0ca62e5dd342a32fe566d61080ba4/examples/basic/main.go#L21-L81

See the [examples][] section for additional use cases.

## Orchestration

The `orchestrator` package provides DAG-based task orchestration on top
of the client library. Define tasks with dependencies, and the library
handles execution order, parallelism, conditional logic, and reporting.

See the [orchestrator example][] for a complete walkthrough.

## Documentation

See the [generated documentation][] for details on available packages and functions.

## Contributing

See the [Development](docs/development.md) guide for prerequisites, setup,
and conventions. See the [Contributing](docs/contributing.md) guide before
submitting a PR.

## License

The [MIT][] License.

[OSAPI]: https://github.com/osapi-io/osapi
[examples]: examples/
[generated documentation]: docs/gen/
[orchestrator example]: examples/orchestrator/main.go
[MIT]: LICENSE
