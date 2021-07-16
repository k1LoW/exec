package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	return killall(cmd) // fallback
}

func killall(cmd *exec.Cmd) error {
	var err error
	wpid := cmd.Process.Pid
	if os.Getenv("TERM") == "cygwin" {
		wpid, err = winpid(cmd.Process.Pid)
		if err != nil {
			return err
		}
	}

	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(wpid)).Run() // #nosec
	// return psutil.TerminateTree(cmd.Process.Pid, 0)
}

// winpid convert cygwin pid -> windows pid
func winpid(pid int) (int, error) {
	winpidCmd := exec.Command("cat", fmt.Sprintf("/proc/%d/winpid", pid)) // #nosec
	out, err := winpidCmd.Output()
	if err != nil {
		out, err = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid)).Output() // #nosec
		if err != nil {
			return pid, err
		}
		if !strings.Contains(string(out), strconv.Itoa(pid)) {
			return pid, errors.New("process does not exist")
		}
		return pid, nil
	}
	winpid, err := strconv.Atoi(strings.TrimRight(string(out), "\n\r"))
	if err != nil {
		return pid, err
	}
	return winpid, nil
}
