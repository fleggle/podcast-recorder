//go:build darwin

package engine

import (
	"os/exec"
	"regexp"
	"strings"
)

func getFFmpegFormat() string {
	return "avfoundation"
}

func getAvailableDevices() ([]Device, error) {
	// ffmpeg -f avfoundation -list_devices true -i ""
	cmd := exec.Command(ffmpegPath(), "-f", "avfoundation", "-list_devices", "true", "-i", "")
	out, err := cmd.CombinedOutput()
	// ffmpeg returns error code when using -list_devices true
	_ = err 

	output := string(out)
	
	// Sample output block we care about:
	// [AVFoundation indev @ 0x147704170] AVFoundation audio devices:
	// [AVFoundation indev @ 0x147704170] [0] Built-in Microphone
	// [AVFoundation indev @ 0x147704170] [1] External Microphone
	
	devices := []Device{}
	inAudioSection := false
	
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`\[(\d+)\]\s+(.*)`)
	
	for _, line := range lines {
		if strings.Contains(line, "AVFoundation audio devices:") {
			inAudioSection = true
			continue
		}
		if inAudioSection {
			// Stop if we hit video section or EOF
			if strings.Contains(line, "AVFoundation video devices:") {
				break
			}
			
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				devices = append(devices, Device{
					ID:   ":" + matches[1], // format for audio only in avfoundation is ":index"
					Name: strings.TrimSpace(matches[2]),
				})
			}
		}
	}
	
	return devices, nil
}
