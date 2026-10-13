package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"podcastrecorder/pkg/engine"
	"podcastrecorder/ui"
)

func main() {
	a := app.NewWithID("com.podcastrecorder.app")
	w := a.NewWindow("Podcast Recorder")
	w.Resize(fyne.NewSize(800, 600))
	
	recorder := engine.NewDesktopFFmpegRecorder()
	
	appUI := ui.NewAppUI(w, recorder)
	
	w.SetOnClosed(func() {
		appUI.StopRecordingGracefully()
	})
	
	w.ShowAndRun()
}
