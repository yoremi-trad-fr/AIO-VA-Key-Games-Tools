package main

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// hideWindow hides the console window on Windows so no CMD popup appears
// during subprocess execution (especially visible during batch image operations).
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
