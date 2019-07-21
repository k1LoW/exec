package exec

import (
	"os"
	"os/exec"
	"syscall"
)

func command(name string, arg ...string) *exec.Cmd {
	// #nosec
	cmd := exec.Command(name, arg...)
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	return cmd
}

func terminate(cmd *exec.Cmd, sig os.Signal) error {
	syssig, ok := sig.(syscall.Signal)
	if !ok {
		return cmd.Process.Signal(sig)
	}
	err := syscall.Kill(-cmd.Process.Pid, syssig)
	if err != nil {
		return err
	}
	if syssig != syscall.SIGKILL && syssig != syscall.SIGCONT {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGCONT)
	}
	return nil
}

func killall(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
