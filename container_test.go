package testingdock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func TestNewContainer_Defaults(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	c := newContainer(t, cli, ContainerOpts{
		Name: "test-container",
		Config: &container.Config{
			Image: "alpine:latest",
		},
	})

	// default HealthCheckTimeout is 30s
	if c.healthchecktimeout != 30*time.Second {
		t.Fatalf("expected default healthchecktimeout of 30s, got %s", c.healthchecktimeout)
	}

	// default HostConfig is non-nil
	if c.hcfg == nil {
		t.Fatal("expected non-nil HostConfig")
	}

	// default Reset func is set
	if c.resetF == nil {
		t.Fatal("expected non-nil resetF")
	}

	// default HealthCheck is set
	if c.healthcheck == nil {
		t.Fatal("expected non-nil healthcheck")
	}

	// labels should be set to testingdock owner
	if !isOwnedByTestingdock(c.ccfg.Labels) {
		t.Fatal("expected testingdock owner label on container config")
	}

	// Name and Image should be set
	if c.Name != "test-container" {
		t.Fatalf("expected Name 'test-container', got %q", c.Name)
	}
	if c.Image != "alpine:latest" {
		t.Fatalf("expected Image 'alpine:latest', got %q", c.Image)
	}
}

func TestNewContainer_CustomTimeout(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	c := newContainer(t, cli, ContainerOpts{
		Name: "test-container",
		Config: &container.Config{
			Image: "alpine:latest",
		},
		HealthCheckTimeout: 60 * time.Second,
	})

	if c.healthchecktimeout != 60*time.Second {
		t.Fatalf("expected healthchecktimeout of 60s, got %s", c.healthchecktimeout)
	}
}

func TestNewContainer_CustomHostConfig(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	hcfg := &container.HostConfig{
		Privileged: true,
	}
	c := newContainer(t, cli, ContainerOpts{
		Name: "test-container",
		Config: &container.Config{
			Image: "alpine:latest",
		},
		HostConfig: hcfg,
	})

	if c.hcfg != hcfg {
		t.Fatal("expected custom HostConfig to be preserved")
	}
}

func TestHealthCheckCustom(t *testing.T) {
	called := false
	fn := func() error {
		called = true
		return nil
	}

	hc := HealthCheckCustom(fn)
	err := hc(context.Background(), &Container{})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !called {
		t.Fatal("expected inner function to be called")
	}
}

func TestHealthCheckCustom_ReturnsError(t *testing.T) {
	expectedErr := errors.New("health check failed")
	fn := func() error {
		return expectedErr
	}

	hc := HealthCheckCustom(fn)
	err := hc(context.Background(), &Container{})
	if err != expectedErr {
		t.Fatalf("expected error %q, got %v", expectedErr, err)
	}
}

func TestResetCustom(t *testing.T) {
	called := false
	fn := func() error {
		called = true
		return nil
	}

	rf := ResetCustom(fn)
	err := rf(context.Background(), &Container{})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !called {
		t.Fatal("expected inner function to be called")
	}
}

func TestResetCustom_ReturnsError(t *testing.T) {
	expectedErr := errors.New("reset failed")
	fn := func() error {
		return expectedErr
	}

	rf := ResetCustom(fn)
	err := rf(context.Background(), &Container{})
	if err != expectedErr {
		t.Fatalf("expected error %q, got %v", expectedErr, err)
	}
}
