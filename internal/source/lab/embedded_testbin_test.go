package lab

import (
	"errors"
	"os"
	"runtime"
	"testing"
)

func TestResolveTestbin(t *testing.T) {
	path, cleanup, err := ResolveTestbin()
	if !supportedPlatform(runtime.GOOS, runtime.GOARCH) {
		if err == nil {
			cleanup()
			t.Fatalf("expected error for unsupported platform")
		}
		if !errors.Is(err, ErrUnsupportedPlatform) {
			t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("resolve testbin: %v", err)
	}
	t.Cleanup(cleanup)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat testbin: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non-empty testbin file")
	}
}
