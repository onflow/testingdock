package testingdock

import (
	"strconv"
	"testing"
)

func TestIsOwnedByTestingdock(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		labels := map[string]string{"owner": "testingdock"}
		if !isOwnedByTestingdock(labels) {
			t.Fatal("expected true for owner=testingdock label")
		}
	})

	t.Run("negative", func(t *testing.T) {
		labels := map[string]string{"owner": "someone-else"}
		if isOwnedByTestingdock(labels) {
			t.Fatal("expected false for owner=someone-else label")
		}
	})

	t.Run("empty labels", func(t *testing.T) {
		labels := map[string]string{}
		if isOwnedByTestingdock(labels) {
			t.Fatal("expected false for empty labels")
		}
	})

	t.Run("nil labels", func(t *testing.T) {
		if isOwnedByTestingdock(nil) {
			t.Fatal("expected false for nil labels")
		}
	})

	t.Run("extra labels present", func(t *testing.T) {
		labels := map[string]string{
			"env":   "test",
			"owner": "testingdock",
			"foo":   "bar",
		}
		if !isOwnedByTestingdock(labels) {
			t.Fatal("expected true when owner=testingdock is among other labels")
		}
	})
}

func TestCreateTestingLabel(t *testing.T) {
	labels := createTestingLabel()

	if labels == nil {
		t.Fatal("expected non-nil labels map")
	}

	val, ok := labels["owner"]
	if !ok {
		t.Fatal("expected 'owner' key in labels")
	}
	if val != "testingdock" {
		t.Fatalf("expected owner=testingdock, got owner=%s", val)
	}
	if len(labels) != 1 {
		t.Fatalf("expected exactly 1 label, got %d", len(labels))
	}
}

func TestRandomPort(t *testing.T) {
	port := RandomPort(t)

	if port == "" {
		t.Fatal("expected non-empty port string")
	}

	num, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("expected numeric port string, got %q: %s", port, err)
	}

	if num < 1 || num > 65535 {
		t.Fatalf("expected port in range 1-65535, got %d", num)
	}
}

func TestRandomPort_Unique(t *testing.T) {
	port1 := RandomPort(t)
	port2 := RandomPort(t)

	// While not guaranteed, two sequential calls should almost always return different ports
	if port1 == port2 {
		t.Logf("warning: two sequential RandomPort calls returned the same port: %s (not necessarily a bug)", port1)
	}
}
