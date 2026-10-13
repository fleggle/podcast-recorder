package engine

import (
	"regexp"
	"strings"
)

// parseDshowAudioDevices extracts audio device names from the combined
// stdout/stderr of `ffmpeg -list_devices true -f dshow -i dummy`.
//
// ffmpeg's dshow device listing format has changed across versions:
//
//   - Older ffmpeg groups devices under "DirectShow audio devices" /
//     "DirectShow video devices" headings, e.g.:
//     [dshow @ ...] DirectShow audio devices
//     [dshow @ ...]  "Microphone (Realtek High Definition Audio)"
//     [dshow @ ...]     Alternative name "@device_cm_{...}"
//
//   - Newer ffmpeg drops the headings and instead tags each device line
//     with "(audio)" or "(video)", e.g.:
//     [in#0 @ ...] "HyperX QuadCast 2" (audio)
//     [in#0 @ ...]   Alternative name "@device_cm_{...}"
//
// Both formats are supported here.
func parseDshowAudioDevices(output string) []Device {
	devices := []Device{}
	inAudioSection := false

	lines := strings.Split(output, "\n")
	reDevice := regexp.MustCompile(`"(.*?)"`)
	reAudioTag := regexp.MustCompile(`\(audio\)\s*$`)
	reVideoTag := regexp.MustCompile(`\(video\)\s*$`)

	for _, line := range lines {
		if strings.Contains(line, "Alternative name") {
			continue
		}

		if strings.Contains(line, "DirectShow audio devices") {
			inAudioSection = true
			continue
		}
		if strings.Contains(line, "DirectShow video devices") {
			inAudioSection = false
			continue
		}

		matches := reDevice.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		name := matches[1]

		isAudio := reAudioTag.MatchString(line)
		isVideo := reVideoTag.MatchString(line)

		switch {
		case isAudio:
			devices = append(devices, Device{ID: "audio=" + name, Name: name})
		case isVideo:
			// explicitly tagged video-only device, skip
		case inAudioSection:
			devices = append(devices, Device{ID: "audio=" + name, Name: name})
		}
	}

	return devices
}
