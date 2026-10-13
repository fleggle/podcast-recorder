//go:build !windows && !android

package engine

import "os/exec"

func hideConsole(cmd *exec.Cmd) {}
