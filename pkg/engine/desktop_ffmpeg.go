//go:build !android

package engine

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type DesktopFFmpegRecorder struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	stdinPipe io.WriteCloser
	recording bool
	doneCh    chan error
}

func NewDesktopFFmpegRecorder() *DesktopFFmpegRecorder {
	return &DesktopFFmpegRecorder{}
}

func (r *DesktopFFmpegRecorder) GetAvailableDevices() ([]Device, error) {
	return getAvailableDevices()
}

func (r *DesktopFFmpegRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
	return reg.ReplaceAllString(name, "_")
}

// BuildFFmpegArgs generates the argument list for FFmpeg given a SessionConfig, timestamp, and format.
// Exported for testing purposes.
func BuildFFmpegArgs(session SessionConfig, timestamp string, format string) []string {
	args := []string{}

	// Build inputs
	for _, track := range session.Tracks {
		args = append(args,
			"-f", format,
			"-use_wallclock_as_timestamps", "1",
			"-i", track.Device.ID,
		)
	}

	// Build outputs
	for i, track := range session.Tracks {
		speaker := sanitizeFilename(track.SpeakerName)
		if speaker == "" {
			speaker = fmt.Sprintf("Track_%d", i+1)
		}

		epName := sanitizeFilename(session.EpisodeName)
		if epName == "" {
			epName = "Episode"
		}

		filename := fmt.Sprintf("%s_%s_%s.wav", epName, timestamp, speaker)
		outPath := filepath.Join(session.OutputDir, filename)

		args = append(args,
			"-map", fmt.Sprintf("%d:a", i),
			"-c:a", "pcm_s24le",
			"-ar", "48000",
			"-ac", "1",
			outPath,
		)
	}

	return args
}

func (r *DesktopFFmpegRecorder) Start(session SessionConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.recording {
		return errors.New("already recording")
	}

	if len(session.Tracks) == 0 {
		return errors.New("no tracks configured")
	}

	format := getFFmpegFormat()
	timestamp := time.Now().Format("20060102_150405")

	args := BuildFFmpegArgs(session, timestamp, format)

	r.cmd = exec.Command(ffmpegPath(), args...)
	hideConsole(r.cmd)
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	r.stdinPipe = stdin

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	r.recording = true
	r.doneCh = make(chan error, 1)

	go func() {
		err := r.cmd.Wait()
		r.mu.Lock()
		r.recording = false
		r.cmd = nil
		r.mu.Unlock()
		r.doneCh <- err
	}()

	return nil
}

func (r *DesktopFFmpegRecorder) Stop() error {
	r.mu.Lock()
	if !r.recording {
		r.mu.Unlock()
		return nil
	}
	stdin := r.stdinPipe
	doneCh := r.doneCh
	r.mu.Unlock()

	if stdin != nil {
		_, err := stdin.Write([]byte("q\n"))
		if err != nil {
			return fmt.Errorf("failed to write to ffmpeg stdin: %w", err)
		}
		stdin.Close()
	}

	<-doneCh

	return nil
}
