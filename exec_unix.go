package exec

import (
	"os"
	"os/exec"
	"syscall"
)

// Reference code:
// https://github.com/Songmu/timeout/blob/9710262dc02f66fdd69a6cd4c8143204006d5843/timeout_unix.go

// Copyright (c) 2015 Songmu
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

func command(name string, arg ...string) *exec.Cmd {
	// #nosec
	cmd := exec.Command(name, arg...)
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
	err := syscall.Kill(-cmd.Process.Pid, syssig)
	if err != nil {
		return syscall.Kill(cmd.Process.Pid, syssig) // fallback
	}
	if syssig != syscall.SIGKILL && syssig != syscall.SIGCONT {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGCONT)
	}
	return nil
}

func killall(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
