# testingdock

A Go library for managing Docker containers and networks in integration tests. Define container dependencies as a tree, and testingdock starts them in parallel with health checks, then tears everything down when the test finishes.

## Install

```bash
go get github.com/onflow/testingdock
```

Requires Go 1.22+ and Docker SDK v28.

## Usage

```go
func TestMain(m *testing.M) {
    flag.Parse()
    code := m.Run()
    testingdock.UnregisterAll()
    os.Exit(code)
}

func TestExample(t *testing.T) {
    // Create a suite (reuses an existing one if the name matches)
    s, _ := testingdock.GetOrCreateSuite(t, "my-suite", testingdock.SuiteOpts{})

    // Create a network
    n := s.Network(testingdock.NetworkOpts{Name: "my-network"})

    // Create containers
    postgresPort := testingdock.RandomPort(t)
    db, _ := sql.Open("postgres", "postgres://postgres:@localhost:"+postgresPort+"?sslmode=disable")

    postgres := s.Container(testingdock.ContainerOpts{
        Name: "postgres",
        Config: &container.Config{
            Image: "postgres:16",
            Env:   []string{"POSTGRES_HOST_AUTH_METHOD=trust"},
        },
        HostConfig: &container.HostConfig{
            PortBindings: nat.PortMap{
                "5432/tcp": []nat.PortBinding{{HostPort: postgresPort}},
            },
        },
        HealthCheck: testingdock.HealthCheckCustom(db.Ping),
    })

    app := s.Container(testingdock.ContainerOpts{
        Name: "app",
        Config: &container.Config{Image: "my-app:latest"},
        HealthCheck: testingdock.HealthCheckHTTP("http://localhost:8080/health"),
    })

    // Define dependency tree: app starts after postgres, both on the network
    n.After(postgres)
    postgres.After(app)

    // Start everything (parallel by default)
    s.Start(context.Background())
    defer s.Close()

    // Run tests...
}
```

## Key Concepts

**Suites** group a network and its containers. Create them with `GetOrCreateSuite` and clean up with `UnregisterAll` in `TestMain`.

**Dependency tree** — `network.After(container)` adds a container to the network. `container.After(child)` makes `child` start after `container`'s health check passes. Siblings at the same level start in parallel by default.

**Health checks** block until the container is ready:
- `HealthCheckHTTP(url)` — polls until HTTP 200
- `HealthCheckCustom(fn)` — calls your `func() error` until it returns nil
- Default — checks the Docker container state is "running"
- Timeout defaults to 30s, configurable via `ContainerOpts.HealthCheckTimeout`

**Reset** re-initializes containers between test cases without full teardown:
```go
s.Reset(ctx) // calls ResetFunc + HealthCheck for each container
```
Default reset restarts the container. Use `ResetCustom(fn)` for custom logic (e.g., truncating database tables).

**Private registry support** — if the image name contains a domain (e.g., `quay.io/org/image:tag`), credentials are read from `~/.docker/config.json` automatically.

## Flags

Pass these via `go test`:

```
-testingdock.sequential    Start containers sequentially instead of in parallel
-testingdock.verbose       Enable verbose logging
```

## Notes

This library creates networks and containers under the label `owner=testingdock`.
Resources with this label are considered owned by testingdock and may be aggressively cleaned up on startup and teardown. Existing containers or networks with the same name that were **not** created by testingdock will cause tests to abort.
