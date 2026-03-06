package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.String()
}

func TestPrintJSON_StableFormatting(t *testing.T) {
	out := captureStdout(t, func() {
		err := PrintJSON(map[string]any{"a": 1, "b": "x"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "\"a\": 1") || !strings.Contains(out, "\"b\": \"x\"") {
		t.Fatalf("unexpected json output: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("expected trailing newline, got %q", out)
	}
}

func TestPrintTable_StableColumns(t *testing.T) {
	out := captureStdout(t, func() {
		PrintTable([]string{"ID", "Name"}, [][]string{{"1", "Checking"}})
	})

	if !strings.Contains(out, "| ID") || !strings.Contains(out, "NAME") {
		t.Fatalf("missing header output: %q", out)
	}
	if !strings.Contains(out, "|  1 |") || !strings.Contains(out, "Checking") {
		t.Fatalf("missing row output: %q", out)
	}
}
