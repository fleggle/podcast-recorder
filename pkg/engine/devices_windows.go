//go:build windows

package engine

import (
	"os/exec"
	"regexp"
	"strings"
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

	output := string(out)
	
	// Sample output:
	// [dshow @ 0000021c3858c9c0] DirectShow audio devices
	// [dshow @ 0000021c3858c9c0]  "Microphone (Realtek High Definition Audio)"
	// [dshow @ 0000021c3858c9c0]     Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{E4F427DE-DB0A-4F7F-9922-95FCFC70ED3C}"
	
	devices := []Device{}
	inAudioSection := false
	
	lines := strings.Split(output, "\n")
	reDevice := regexp.MustCompile(`\"(.*?)\"`)
	
	var currentName string
	for _, line := range lines {
		if strings.Contains(line, "DirectShow audio devices") {
			inAudioSection = true
			continue
		}
		if inAudioSection {
			if strings.Contains(line, "DirectShow video devices") {
				break
			}
			
			matches := reDevice.FindStringSubmatch(line)
			if len(matches) > 1 {
				val := matches[1]
				if strings.Contains(line, "Alternative name") {
					// We use alternative name if available, otherwise just use audio="Name"
					if currentName != "" {
						devices = append(devices, Device{
							ID:   "audio=" + currentName,
							Name: currentName,
						})
						currentName = ""
					}
				} else {
					if currentName != "" {
						devices = append(devices, Device{
							ID:   "audio=" + currentName,
							Name: currentName,
						})
					}
					currentName = val
				}
			}
		}
	}
	
	if currentName != "" {
		devices = append(devices, Device{
			ID:   "audio=" + currentName,
			Name: currentName,
		})
	}
	
	return devices, nil
}
