package lab

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var ErrUnsupportedPlatform = errors.New("unsupported platform")

//go:embed testbin/*/testbin*
var testbinFS embed.FS

func ResolveTestbin() (string, func(), error) {
	name, err := embeddedTestbinName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", nil, err
	}
	data, err := testbinFS.ReadFile(name)
	if err != nil {
		return "", nil, err
	}

	// Note: prefix is part of the user-visible temp dir name.
	dir, err := os.MkdirTemp("", "gcscope-testbin-*")
	if err != nil {
		return "", nil, err
	}

	outPath := filepath.Join(dir, filepath.Base(name))
	if err := os.WriteFile(outPath, data, 0755); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}

	cleanup := func() { _ = os.RemoveAll(dir) }
	return outPath, cleanup, nil
}

func embeddedTestbinName(goos, goarch string) (string, error) {
	switch {
	case goos == "windows" && goarch == "amd64":
		return "testbin/windows_amd64/testbin.exe", nil
	case goos == "windows" && goarch == "arm64":
		return "testbin/windows_arm64/testbin.exe", nil
	case goos == "linux" && goarch == "amd64":
		return "testbin/linux_amd64/testbin", nil
	case goos == "linux" && goarch == "arm64":
		return "testbin/linux_arm64/testbin", nil
	case goos == "darwin" && goarch == "amd64":
		return "testbin/darwin_amd64/testbin", nil
	case goos == "darwin" && goarch == "arm64":
		return "testbin/darwin_arm64/testbin", nil
	default:
		return "", fmt.Errorf("%w: %s/%s", ErrUnsupportedPlatform, goos, goarch)
	}
}

func supportedPlatform(goos, goarch string) bool {
	_, err := embeddedTestbinName(goos, goarch)
	return err == nil
}
