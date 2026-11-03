//go:build windows

package localai

import (
	"os/exec"
	"strconv"
	"syscall"
)

func setSidecarSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func killProcessTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	_ = exec.Command("taskkill", "/F", "/T", "/PID", pid).Run()
	_, _ = cmd.Process.Wait()
}
