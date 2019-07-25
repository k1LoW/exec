package exec

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// MEMO: Sending Interrupt on Windows is not implemented.
var defaultSignal = os.Interrupt

// Reference code:
// https://github.com/Songmu/timeout/blob/517fff103abc7d0e88a663609515d8bb55f6535d/timeout_windows.go
func command(name string, arg ...string) *exec.Cmd {
	// #nosec
	cmd := exec.Command(name, arg...)
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_UNICODE_ENVIRONMENT | 0x00000200
	return cmd
}

func terminate(cmd *exec.Cmd, sig os.Signal) error {
	if err := cmd.Process.Signal(sig); err != nil {
		return killall(cmd) // fallback
	}
	return nil
}

func killall(cmd *exec.Cmd) error {
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}
