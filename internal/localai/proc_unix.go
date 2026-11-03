//go:build unix

package localai

import (
	"os/exec"
	"syscall"
)

func setSidecarSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	// Kill the whole process group so llama-server cannot outlive the app.
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_, _ = cmd.Process.Wait()
}
