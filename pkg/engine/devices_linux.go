//go:build linux && !android

package engine

import (
	"os/exec"
	"regexp"
	"strings"
)

func getFFmpegFormat() string {
	return "alsa"
}

func getAvailableDevices() ([]Device, error) {
	// try arecord -l first
	cmd := exec.Command("arecord", "-l")
	out, err := cmd.CombinedOutput()
	
	devices := []Device{}
	if err == nil {
		output := string(out)
		lines := strings.Split(output, "\n")
		// card 1: Generic [HD-Audio Generic], device 0: ALC897 Analog [ALC897 Analog]
		re := regexp.MustCompile(`card (\d+):.*device (\d+):`)
		nameRe := regexp.MustCompile(`card \d+: (.*?) \[`)
		
		for _, line := range lines {
			if strings.HasPrefix(line, "card") {
				matches := re.FindStringSubmatch(line)
				if len(matches) == 3 {
					card := matches[1]
					dev := matches[2]
					
					name := ""
					nameMatch := nameRe.FindStringSubmatch(line)
					if len(nameMatch) > 1 {
						name = nameMatch[1]
					}
					
					devices = append(devices, Device{
						ID:   "hw:" + card + "," + dev,
						Name: name,
					})
				}
			}
		}
		
		if len(devices) > 0 {
			return devices, nil
		}
	}
	
	// Fallback to /proc/asound/cards
	// 1 [Generic        ]: HDA-Intel - HD-Audio Generic
	out, err = exec.Command("cat", "/proc/asound/cards").CombinedOutput()
	if err == nil {
		output := string(out)
		lines := strings.Split(output, "\n")
		re := regexp.MustCompile(`^\s*(\d+)\s*\[.*?\]:\s*(.*)`)
		for _, line := range lines {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 2 {
				devices = append(devices, Device{
					ID:   "hw:" + matches[1] + ",0",
					Name: strings.TrimSpace(matches[2]),
				})
			}
		}
	}
	
	// If still no devices, return a default one so it doesn't panic
	if len(devices) == 0 {
		devices = append(devices, Device{
			ID:   "default",
			Name: "Default Audio Device",
		})
	}
	
	return devices, nil
}
