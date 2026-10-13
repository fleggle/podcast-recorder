package engine

import (
	"reflect"
	"testing"
)

func TestBuildFFmpegArgs(t *testing.T) {
	tests := []struct {
		name      string
		session   SessionConfig
		timestamp string
		format    string
		expected  []string
	}{
		{
			name: "1 track",
			session: SessionConfig{
				EpisodeName: "TestEpisode",
				OutputDir:   "/out",
				Tracks: []TrackConfig{
					{SpeakerName: "Host", Device: Device{ID: "hw:0,0"}},
				},
			},
			timestamp: "20230101_120000",
			format:    "alsa",
			expected: []string{
				"-f", "alsa", "-use_wallclock_as_timestamps", "1", "-i", "hw:0,0",
				"-map", "0:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "/out/TestEpisode_20230101_120000_Host.wav",
			},
		},
		{
			name: "2 tracks",
			session: SessionConfig{
				EpisodeName: "My_Podcast",
				OutputDir:   "/data/podcasts",
				Tracks: []TrackConfig{
					{SpeakerName: "Host 1", Device: Device{ID: ":0"}},
					{SpeakerName: "Guest!@#", Device: Device{ID: ":1"}},
				},
			},
			timestamp: "20230102_130000",
			format:    "avfoundation",
			expected: []string{
				"-f", "avfoundation", "-use_wallclock_as_timestamps", "1", "-i", ":0",
				"-f", "avfoundation", "-use_wallclock_as_timestamps", "1", "-i", ":1",
				"-map", "0:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "/data/podcasts/My_Podcast_20230102_130000_Host_1.wav",
				"-map", "1:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "/data/podcasts/My_Podcast_20230102_130000_Guest___.wav",
			},
		},
		{
			name: "N tracks with empty names",
			session: SessionConfig{
				EpisodeName: "",
				OutputDir:   "out_dir",
				Tracks: []TrackConfig{
					{SpeakerName: "", Device: Device{ID: "audio=Mic1"}},
					{SpeakerName: "", Device: Device{ID: "audio=Mic2"}},
					{SpeakerName: "", Device: Device{ID: "audio=Mic3"}},
				},
			},
			timestamp: "ts",
			format:    "dshow",
			expected: []string{
				"-f", "dshow", "-use_wallclock_as_timestamps", "1", "-i", "audio=Mic1",
				"-f", "dshow", "-use_wallclock_as_timestamps", "1", "-i", "audio=Mic2",
				"-f", "dshow", "-use_wallclock_as_timestamps", "1", "-i", "audio=Mic3",
				"-map", "0:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "out_dir/Episode_ts_Track_1.wav",
				"-map", "1:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "out_dir/Episode_ts_Track_2.wav",
				"-map", "2:a", "-c:a", "pcm_s24le", "-ar", "48000", "-ac", "1", "out_dir/Episode_ts_Track_3.wav",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := BuildFFmpegArgs(tt.session, tt.timestamp, tt.format)
			if !reflect.DeepEqual(args, tt.expected) {
				t.Errorf("BuildFFmpegArgs() \ngot = %v\nwant = %v", args, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Host", "Host"},
		{"Guest 1", "Guest_1"},
		{"My_Episode-1.0", "My_Episode-1.0"},
		{"Invalid<>Characters/\\|?:*", "Invalid__Characters______"},
		{"Hello!@#World", "Hello___World"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := sanitizeFilename(tt.input); got != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
