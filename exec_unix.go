// +build !windows

package exec

import (
	"os"
	"os/exec"
	"syscall"
)

var defaultSignal = syscall.SIGTERM

// Reference code:
// https://github.com/Songmu/timeout/blob/9710262dc02f66fdd69a6cd4c8143204006d5843/timeout_unix.go
func command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...) // #nosec
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true // force setpgid
	return cmd
}

func terminate(cmd *exec.Cmd, sig os.Signal) error {
	syssig, ok := sig.(syscall.Signal)
	if !ok {
		return cmd.Process.Signal(sig)
	}
	if cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syssig)
	if err != nil {
		return syscall.Kill(cmd.Process.Pid, syssig) // fallback
	}
	if syssig != syscall.SIGKILL && syssig != syscall.SIGCONT {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGCONT)
	}
	return nil
}

func killall(cmd *exec.Cmd) error {
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		return cmd.Process.Kill() // fallback
	}
	return nil
}
