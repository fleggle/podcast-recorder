package engine

// Device represents an audio input device
type Device struct {
	ID   string // e.g., "hw:1,0", "audio=MicName", "none:0"
	Name string // Human-readable display name
}

// TrackConfig defines the configuration for a single recording track
type TrackConfig struct {
	SpeakerName string
	Device      Device
}

// SessionConfig defines the configuration for a recording session
type SessionConfig struct {
	EpisodeName string
	OutputDir   string
	Tracks      []TrackConfig
}

// AudioRecorder is the interface for the audio recording engine
type AudioRecorder interface {
	GetAvailableDevices() ([]Device, error)
	Start(session SessionConfig) error
	Stop() error
	IsRecording() bool
}
