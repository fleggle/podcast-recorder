package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"podcastrecorder/pkg/engine"
)

type AppUI struct {
	window       fyne.Window
	recorder     engine.AudioRecorder
	devices      []engine.Device
	
	mu           sync.Mutex
	tracks       []*TrackUI
	isRecording  bool
	
	// UI Elements
	btnRecord      *widget.Button
	btnAddMic      *widget.Button
	btnRefresh     *widget.Button
	btnBrowse      *widget.Button
	
	entryEpisode   *widget.Entry
	labelTimer     *widget.Label
	
	trackList      *fyne.Container
	
	lblState       *widget.Label
	lblPath        *widget.Label
	lblStats       *widget.Label
	
	currentPath    string
	recordingTimer *time.Timer
	timerTicker    *time.Ticker
	startTime      time.Time
}

type TrackUI struct {
	container *fyne.Container
	speaker   *widget.Entry
	deviceSel *widget.Select
	subtext   *widget.Label
	btnTrash  *widget.Button
	devices   []engine.Device
}

func NewAppUI(win fyne.Window, rec engine.AudioRecorder) *AppUI {
	ui := &AppUI{
		window:   win,
		recorder: rec,
		tracks:   make([]*TrackUI, 0),
		currentPath: "./", // default
	}
	
	ui.buildUI()
	ui.refreshDevices()
	ui.addTrack() // Start with 1 track
	
	return ui
}

func (ui *AppUI) buildUI() {
	// Top Toolbar
	ui.btnRecord = widget.NewButtonWithIcon("Record", theme.MediaRecordIcon(), ui.toggleRecording)
	ui.btnAddMic = widget.NewButtonWithIcon("Add Mic", theme.ContentAddIcon(), ui.addTrack)
	ui.btnRefresh = widget.NewButtonWithIcon("Refresh Devices", theme.ViewRefreshIcon(), ui.refreshDevices)
	ui.btnBrowse = widget.NewButtonWithIcon("Output Folder", theme.FolderIcon(), ui.browseFolder)
	
	toolbar := container.NewHBox(
		ui.btnRecord,
		ui.btnAddMic,
		ui.btnRefresh,
		ui.btnBrowse,
	)
	
	// Sub-bar
	ui.entryEpisode = widget.NewEntry()
	ui.entryEpisode.SetPlaceHolder("Episode Name...")
	
	ui.labelTimer = widget.NewLabel("00:00:00")
	ui.labelTimer.Alignment = fyne.TextAlignTrailing
	ui.labelTimer.TextStyle = fyne.TextStyle{Monospace: true}
	
	subbar := container.NewBorder(nil, nil, nil, ui.labelTimer, ui.entryEpisode)
	
	topSection := container.NewVBox(toolbar, subbar)
	
	// Central Track List
	ui.trackList = container.NewVBox()
	scrollList := container.NewVScroll(ui.trackList)
	
	// Bottom Status Strip
	ui.lblState = widget.NewLabel("Idle")
	ui.lblPath = widget.NewLabel(ui.currentPath)
	ui.lblStats = widget.NewLabel("0 tracks • 48 kHz")
	ui.lblStats.Alignment = fyne.TextAlignTrailing
	
	statusStrip := container.NewBorder(nil, nil, ui.lblState, ui.lblStats, ui.lblPath)
	
	content := container.NewBorder(topSection, statusStrip, nil, nil, scrollList)
	ui.window.SetContent(content)
}

func (ui *AppUI) refreshDevices() {
	devs, err := ui.recorder.GetAvailableDevices()
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}
	ui.mu.Lock()
	ui.devices = devs
	ui.mu.Unlock()
	
	deviceNames := []string{}
	for _, d := range devs {
		deviceNames = append(deviceNames, d.Name)
	}
	
	ui.mu.Lock()
	for _, t := range ui.tracks {
		t.devices = devs
		t.deviceSel.Options = deviceNames
		t.deviceSel.Refresh()
	}
	ui.mu.Unlock()
}

func (ui *AppUI) updateStats() {
	ui.lblStats.SetText(fmt.Sprintf("%d tracks • 48 kHz", len(ui.tracks)))
}

func (ui *AppUI) addTrack() {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	
	if ui.isRecording {
		return
	}
	
	trackIndex := len(ui.tracks)
	t := &TrackUI{
		devices: ui.devices,
	}
	
	t.speaker = widget.NewEntry()
	t.speaker.SetPlaceHolder(fmt.Sprintf("Guest %d", trackIndex+1))
	
	deviceNames := []string{}
	for _, d := range ui.devices {
		deviceNames = append(deviceNames, d.Name)
	}
	t.deviceSel = widget.NewSelect(deviceNames, func(s string) {
		ui.updateTrackSubtext(t)
	})
	t.subtext = widget.NewLabel("PCM 24-bit • 48 kHz • Mono | None")
	t.subtext.TextStyle = fyne.TextStyle{Italic: true}
	if len(deviceNames) > 0 {
		t.deviceSel.SetSelectedIndex(0)
	}
	
	t.btnTrash = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		ui.removeTrack(t)
	})
	
	topLine := container.NewBorder(nil, nil, nil, nil, container.NewGridWithColumns(2, t.speaker, t.deviceSel))
	cardContent := container.NewVBox(topLine, t.subtext)
	t.container = container.NewBorder(nil, nil, nil, t.btnTrash, cardContent)
	
	ui.tracks = append(ui.tracks, t)
	ui.trackList.Add(t.container)
	ui.trackList.Refresh()
	
	ui.updateTrackSubtext(t)
	ui.updateStats()
}

func (ui *AppUI) removeTrack(t *TrackUI) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	
	if ui.isRecording {
		return
	}
	
	for i, track := range ui.tracks {
		if track == t {
			ui.tracks = append(ui.tracks[:i], ui.tracks[i+1:]...)
			ui.trackList.Remove(t.container)
			ui.trackList.Refresh()
			break
		}
	}
	ui.updateStats()
}

func (ui *AppUI) updateTrackSubtext(t *TrackUI) {
	selName := t.deviceSel.Selected
	devID := "Unknown"
	for _, d := range t.devices {
		if d.Name == selName {
			devID = d.ID
			break
		}
	}
	
	ep := ui.entryEpisode.Text
	if ep == "" {
		ep = "Episode"
	}
	spk := t.speaker.Text
	if spk == "" {
		spk = "Track"
	}
	
	// A bit hacky to show a preview since we don't have timestamp yet
	previewFile := fmt.Sprintf("%s_YYYYMMDD_HHMMSS_%s.wav", ep, spk)
	t.subtext.SetText(fmt.Sprintf("PCM 24-bit • 48 kHz • Mono | ID: %s | %s", devID, previewFile))
}

func (ui *AppUI) browseFolder() {
	dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
		if err == nil && lu != nil {
			ui.mu.Lock()
			ui.currentPath = lu.Path()
			ui.mu.Unlock()
			ui.lblPath.SetText(ui.currentPath)
		}
	}, ui.window)
}

func (ui *AppUI) toggleRecording() {
	ui.mu.Lock()
	isRec := ui.isRecording
	ui.mu.Unlock()
	
	if isRec {
		ui.stopRecording()
	} else {
		ui.startRecording()
	}
}

func (ui *AppUI) startRecording() {
	ui.mu.Lock()
	
	trackConfigs := []engine.TrackConfig{}
	for _, t := range ui.tracks {
		selName := t.deviceSel.Selected
		var dev engine.Device
		for _, d := range ui.devices {
			if d.Name == selName {
				dev = d
				break
			}
		}
		
		spk := t.speaker.Text
		if spk == "" {
			spk = t.speaker.PlaceHolder
		}
		
		trackConfigs = append(trackConfigs, engine.TrackConfig{
			SpeakerName: spk,
			Device:      dev,
		})
	}
	
	session := engine.SessionConfig{
		EpisodeName: ui.entryEpisode.Text,
		OutputDir:   ui.currentPath,
		Tracks:      trackConfigs,
	}
	ui.mu.Unlock()
	
	err := ui.recorder.Start(session)
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}
	
	ui.mu.Lock()
	ui.isRecording = true
	ui.startTime = time.Now()
	
	// Update UI
	ui.btnRecord.SetText("Stop")
	ui.btnRecord.SetIcon(theme.MediaStopIcon())
	ui.btnAddMic.Disable()
	ui.entryEpisode.Disable()
	for _, t := range ui.tracks {
		t.btnTrash.Disable()
		t.deviceSel.Disable()
		t.speaker.Disable()
	}
	
	ui.lblState.SetText(fmt.Sprintf("Recording %d tracks...", len(ui.tracks)))
	ui.mu.Unlock()
	
	ui.timerTicker = time.NewTicker(time.Second)
	go func() {
		for range ui.timerTicker.C {
			ui.mu.Lock()
			if !ui.isRecording {
				ui.mu.Unlock()
				break
			}
			dur := time.Since(ui.startTime)
			dStr := fmt.Sprintf("%02d:%02d:%02d", int(dur.Hours()), int(dur.Minutes())%60, int(dur.Seconds())%60)
			ui.mu.Unlock()
			
			fyne.Do(func() {
				ui.labelTimer.SetText(dStr)
			})
		}
	}()
}

func (ui *AppUI) stopRecording() {
	ui.mu.Lock()
	ui.isRecording = false
	if ui.timerTicker != nil {
		ui.timerTicker.Stop()
	}
	ui.lblState.SetText("Flushing WAV headers...")
	ui.mu.Unlock()
	
	go func() {
		err := ui.recorder.Stop()
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, ui.window)
			}
			
			ui.mu.Lock()
			ui.btnRecord.SetText("Record")
			ui.btnRecord.SetIcon(theme.MediaRecordIcon())
			ui.btnAddMic.Enable()
			ui.entryEpisode.Enable()
			for _, t := range ui.tracks {
				t.btnTrash.Enable()
				t.deviceSel.Enable()
				t.speaker.Enable()
			}
			ui.lblState.SetText("Idle")
			ui.mu.Unlock()
		})
	}()
}

func (ui *AppUI) StopRecordingGracefully() {
	ui.mu.Lock()
	isRec := ui.isRecording
	ui.mu.Unlock()
	
	if isRec {
		ui.recorder.Stop()
	}
}
