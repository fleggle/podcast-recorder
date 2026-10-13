//go:build windows

package engine

import (
	"os/exec"
)

func getFFmpegFormat() string {
	return "dshow"
}

func getAvailableDevices() ([]Device, error) {
	// ffmpeg -list_devices true -f dshow -i dummy
	cmd := exec.Command(ffmpegPath(), "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	hideConsole(cmd)
	out, err := cmd.CombinedOutput()
	_ = err

	return parseDshowAudioDevices(string(out)), nil
}
