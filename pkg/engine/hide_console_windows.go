//go:build windows

package engine

import (
	"os/exec"
	"syscall"
)

// hideConsole stops ffmpeg from opening a console window when launched from the GUI app.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
