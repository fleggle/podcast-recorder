package engine

import (
	"reflect"
	"testing"
)

func TestParseDshowAudioDevices_OldHeadingFormat(t *testing.T) {
	output := `[dshow @ 0000021c3858c9c0] DirectShow video devices
[dshow @ 0000021c3858c9c0]  "HP 5MP Camera"
[dshow @ 0000021c3858c9c0]     Alternative name "@device_pnp_\\?\usb#vid_04f2&pid_b7e9&mi_00"
[dshow @ 0000021c3858c9c0] DirectShow audio devices
[dshow @ 0000021c3858c9c0]  "Microphone (Realtek High Definition Audio)"
[dshow @ 0000021c3858c9c0]     Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{BBBB}"
`
	want := []Device{
		{ID: "audio=Microphone (Realtek High Definition Audio)", Name: "Microphone (Realtek High Definition Audio)"},
	}

	got := parseDshowAudioDevices(output)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDshowAudioDevices() =\n got  %#v\n want %#v", got, want)
	}
}

func TestParseDshowAudioDevices_NewTaggedFormat(t *testing.T) {
	output := `[in#0 @ 000001c3d2dc31c0] "HP 5MP Camera" (video)
[in#0 @ 000001c3d2dc31c0]   Alternative name "@device_pnp_\\?\usb#vid_04f2&pid_b7e9&mi_00#7&6a157a8&0&0000#{65e8773d-8f56-11d0-a3b9-00a0c9223196}\global"
[in#0 @ 000001c3d2dc31c0] "HyperX QuadCast 2" (audio)
[in#0 @ 000001c3d2dc31c0]   Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{AAAA}"
[in#0 @ 000001c3d2dc31c0] "Microphone (Realtek High Definition Audio)" (audio)
[in#0 @ 000001c3d2dc31c0]   Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{BBBB}"
`
	want := []Device{
		{ID: "audio=HyperX QuadCast 2", Name: "HyperX QuadCast 2"},
		{ID: "audio=Microphone (Realtek High Definition Audio)", Name: "Microphone (Realtek High Definition Audio)"},
	}

	got := parseDshowAudioDevices(output)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDshowAudioDevices() =\n got  %#v\n want %#v", got, want)
	}
}
