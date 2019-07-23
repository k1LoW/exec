package exec

import (
	"context"
	"os"
	"os/exec"
	"time"
)

type Exec struct {
	Signal          os.Signal
	KillAfterCancel time.Duration // TODO
}

func (e *Exec) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	if e.Signal == nil {
		e.Signal = defaultSignal
	}
	cmd := command(name, arg...)
	go func() {
		select {
		case <-ctx.Done():
			err := terminate(cmd, e.Signal)
			if err != nil {
				// :thinking:
				return
			}
		}
	}()
	return cmd
}

func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func Command(name string, arg ...string) *exec.Cmd {
	return command(name, arg...)
}

func CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	e := &Exec{
		Signal:          os.Kill, // Why os.Kill ? => for get close to the behavior of os/exec.ContextCommand
		KillAfterCancel: -1,
	}
	return e.CommandContext(ctx, name, arg...)
}

func TerminateCommand(cmd *exec.Cmd, sig os.Signal) error {
	return terminate(cmd, sig)
}

func KillCommand(cmd *exec.Cmd) error {
	return killall(cmd)
}
