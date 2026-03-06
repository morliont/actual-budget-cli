package bridge

import (
	"os"
	"testing"
	"time"
)

func TestTimeoutFromEnv(t *testing.T) {
	t.Setenv(timeoutEnvVar, "")
	d, err := timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != defaultTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultTimeout, d)
	}

	t.Setenv(timeoutEnvVar, "45s")
	d, err = timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error for duration: %v", err)
	}
	if d != 45*time.Second {
		t.Fatalf("expected 45s, got %s", d)
	}

	t.Setenv(timeoutEnvVar, "12")
	d, err = timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error for seconds: %v", err)
	}
	if d != 12*time.Second {
		t.Fatalf("expected 12s, got %s", d)
	}

	t.Setenv(timeoutEnvVar, "0")
	if _, err := timeoutFromEnv(); err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestMaterializeBridgeScript(t *testing.T) {
	path, cleanup, err := materializeBridgeScript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("script path should exist: %v", err)
	}
	if st.Size() == 0 {
		t.Fatal("embedded script should not be empty")
	}
}
