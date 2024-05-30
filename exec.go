package exec

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// Exec represents an command executer
type Exec struct {
	Signal          os.Signal
	KillAfterCancel time.Duration // TODO
}

// CommandContext returns *os/exec.Cmd with Setpgid = true
// When ctx cancelled, `github.com/k1LoW/exec.CommandContext` send signal to process group
func (e *Exec) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	if e.Signal == nil {
		e.Signal = defaultSignal
	}
	cmd := commandContext(ctx, name, arg...)
	cmd.Cancel = func() error {
		return terminate(cmd, e.Signal)
	}
	return cmd
}

// LookPath is os/exec.LookPath
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Command returns *os/exec.Cmd with Setpgid = true
func Command(name string, arg ...string) *exec.Cmd {
	return command(name, arg...)
}

// CommandContext returns *os/exec.Cmd with Setpgid = true
// When ctx cancelled, `github.com/k1LoW/exec.CommandContext` send signal to process group
func CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	e := &Exec{
		Signal:          os.Kill, // Why os.Kill ? => for get close to the behavior of os/exec.ContextCommand
		KillAfterCancel: -1,
	}
	return e.CommandContext(ctx, name, arg...)
}

// TerminateCommand send signal to cmd.Process.Pid process group ( if runtime.GOOS != 'windows' )
// TerminateCommand send taskkill to cmd.Process.Pid ( if runtime.GOOS == 'windows' )
func TerminateCommand(cmd *exec.Cmd, sig os.Signal) error {
	return terminate(cmd, sig)
}

// KillCommand send syscall.SIGKILL to cmd.Process.Pid process group
// KillCommand send taskkill to cmd.Process.Pid ( if runtime.GOOS == 'windows' )
func KillCommand(cmd *exec.Cmd) error {
	return killall(cmd)
}
