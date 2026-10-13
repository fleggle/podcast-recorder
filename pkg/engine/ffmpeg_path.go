//go:build !android

package engine

import (
	"os"
	"path/filepath"
	"runtime"
)

// ffmpegPath returns an ffmpeg binary bundled next to the executable if present,
// otherwise "ffmpeg" so it is resolved from PATH.
func ffmpegPath() string {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	exe, err := os.Executable()
	if err != nil {
		return "ffmpeg"
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	bundled := filepath.Join(filepath.Dir(exe), name)
	if info, err := os.Stat(bundled); err == nil && !info.IsDir() {
		return bundled
	}
	return "ffmpeg"
}
