# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Codex, Cursor, Copilot, and others)
when working in this repository. It is loaded into agent context automatically — keep it concise.

## Overview

`testingdock` is a Go library that manages Docker containers and networks for integration tests
(`github.com/onflow/testingdock`, see `go.mod`). Tests declare a tree of container dependencies
rooted on a network; the suite starts them (in parallel by default), blocks on per-container
health checks, and tears everything down on `Close`. Requires Go 1.24.0 (`go.mod`) and the
Docker Engine SDK v28.5.2 (`github.com/docker/docker`, `github.com/docker/cli`).

## Build and Test Commands

Go toolchain is invoked with `GO111MODULE=on` via the `Makefile`:

- `make build` — `go build` of the package.
- `make test` — `go test` (requires a running Docker daemon; tests pull real images such as
  `postgres` — see `container_functional_test.go`).
- `make lint` — runs `gometalinter` (declared in `Makefile`; not vendored, must be installed
  separately).

Test flags, registered in `suite.go`'s `init()`:

- `-testingdock.sequential` — spawn sibling containers sequentially instead of in parallel
  (sets `SpawnSequential`; useful for debugging).
- `-testingdock.verbose` — enable verbose daemon log forwarding (sets `Verbose`).

Pass through `go test`, e.g. `go test -testingdock.verbose -run TestContainer_Start`.

## Architecture

Flat package at the repo root (`package testingdock`). Four source files:

- `suite.go` — `Suite`, `SuiteOpts`, `GetOrCreateSuite`, `UnregisterAll`, and the global
  flag registration. Suites are keyed by name in an internal `registry` map.
- `network.go` — `Network`, `NetworkOpts`, startup/reset/remove logic, and initial cleanup
  of stale `owner=testingdock` resources.
- `container.go` — `Container`, `ContainerOpts`, `HealthCheckFunc`/`HealthCheckHTTP`/
  `HealthCheckCustom`, `ResetFunc`/`ResetCustom`, image pulling, and private-registry auth
  via `docker/cli/cli/config` (reads `~/.docker/config.json`).
- `helpers.go` — `RandomPort`, logging `printf`, and the `owner=testingdock` label helpers.

Tests live alongside sources as `*_test.go` (external `testingdock_test` package). A
`TestMain` is required to call `flag.Parse()` and `testingdock.UnregisterAll()` — see
`suite_test.go` for the canonical pattern.

Dependency wiring is a tree: `network.After(container)` attaches a container to the
network; `container.After(child)` makes `child` start only after the parent's health check
passes. Siblings start in parallel unless `-testingdock.sequential` is set (`suite.go`,
`network.go`).

## Conventions and Gotchas

- **Label ownership.** All networks and containers are created with label `owner=testingdock`
  (`helpers.go` `createTestingLabel`). On startup, pre-existing resources with this label are
  aggressively removed; resources with the same name but without the label cause `t.Fatalf`.
  Never remove or rename this label.
- **`TestMain` is mandatory.** Consumers must call `flag.Parse()` then `testingdock.UnregisterAll()`
  in `TestMain`; otherwise test flags aren't bound and containers leak between runs
  (`suite_test.go:12-20`, package doc in `suite.go`).
- **Health check timeout** defaults to 30s when `ContainerOpts.HealthCheckTimeout` is zero
  (`container.go`'s `ContainerOpts` comment). The default (zero-value) health check only
  verifies Docker reports the container as running — override with `HealthCheckHTTP` or
  `HealthCheckCustom` for real readiness.
- **Reset vs. Close.** `Suite.Reset(ctx)` re-runs each container's `ResetFunc` (default:
  `ContainerRestart`) and re-checks health; use `ResetCustom` for in-place cleanup like
  truncating tables. `Suite.Close` tears the network down.
- **Private registries.** If an image reference contains a registry domain, `container.go`
  base64-encodes credentials from `~/.docker/config.json` via `clicfg` for the pull.
- **Port allocation.** Use `testingdock.RandomPort(t)` (`helpers.go`) for host-side port
  bindings; it reserves a free TCP port on `localhost:0` and returns it as a string.
- **Docker daemon required.** Tests instantiate a real client via `client.FromEnv`
  (`suite.go`). Without a reachable daemon they `t.Fatalf` unless `SuiteOpts.Skip` is set.

## Files Not to Modify

- `go.sum` — regenerate via `go mod tidy`, do not hand-edit.
- `LICENSE.txt` — Apache-2.0 header, do not edit.
