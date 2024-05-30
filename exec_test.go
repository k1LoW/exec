package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	shellcmd  = `/bin/sh`
	shellflag = "-c"
	stubCmd   = `./testdata/stubcmd`
)

func init() {
	if runtime.GOOS == "windows" {
		shellcmd = "cmd"
		shellflag = "/c"
		stubCmd = `.\testdata\stubcmd.exe`
	}
	err := exec.Command("go", "build", "-o", stubCmd, "testdata/stubcmd.go").Run()
	if err != nil {
		panic(err)
	}
}

func TestCommand(t *testing.T) {
	tests := gentests(false)
	for _, tt := range tests {
		_ = killprocess()
		if checkprocess() {
			t.Fatalf("%s", "the process has not exited")
		}

		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Errorf("%v: %v", tt.cmd, err)
		}
		if strings.TrimRight(stdout.String(), "\n\r") != tt.want {
			t.Errorf("%v: want = %#v, got = %#v", tt.cmd, tt.want, stdout.String())
		}
		if checkprocess() {
			t.Errorf("%v: %s", tt.cmd, "the process has not exited")
		}
	}
}

func TestCommandContext(t *testing.T) {
	tests := gentests(false)
	for _, tt := range tests {
		_ = killprocess()
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		ctx := context.Background()
		cmd := CommandContext(ctx, tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Fatalf("%v: %v", tt.cmd, err)
		}
		if strings.TrimRight(stdout.String(), "\n\r") != tt.want {
			t.Errorf("%v: want = %#v, got = %#v", tt.cmd, tt.want, stdout.String())
		}
		if checkprocess() {
			t.Errorf("%v: %s", tt.cmd, "the process has not exited")
		}
	}
}

func TestCommandContextCancel(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		_ = killprocess()
		if checkprocess() {
			t.Fatalf("%s", "the process has not exited")
		}

		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		ctx, cancel := context.WithCancel(context.Background())
		cmd := CommandContext(ctx, tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			cancel()
			t.Fatalf("%v: %v", tt.cmd, err)
		}
		go func() {
			cmd.Wait()
		}()
		if !checkprocess() && !tt.processFinished {
			cancel()
			t.Fatalf("%v: %s", tt.cmd, "the process has been exited")
		}
		cancel()
		if checkprocess() {
			t.Errorf("%v: %s", tt.cmd, "the process has not exited")
		}
	}
}

func TestTerminateCommand(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		_ = killprocess()
		if checkprocess() {
			t.Fatalf("%s", "the process has not exited")
		}

		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			t.Fatalf("%v: %v", tt.cmd, err)
		}
		go func() {
			cmd.Wait()
		}()
		if !checkprocess() && !tt.processFinished {
			t.Fatalf("%v: %s", tt.cmd, "the process has been exited")
		}
		if runtime.GOOS == "windows" {
			sig := os.Interrupt
			err = TerminateCommand(cmd, sig)
		} else {
			sig := syscall.SIGTERM
			err = TerminateCommand(cmd, sig)
		}
		if err != nil && !tt.processFinished {
			t.Errorf("%v: %v", tt.cmd, err)
		}
		if checkprocess() {
			t.Errorf("%v: %s", tt.cmd, "the process has not exited")
		}
	}
}

func TestKillCommand(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		_ = killprocess()
		if checkprocess() {
			t.Fatalf("%s", "the process has not exited")
		}

		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			t.Fatalf("%v: %v", tt.cmd, err)
		}
		go func() {
			cmd.Wait()
		}()
		if !checkprocess() && !tt.processFinished {
			t.Fatalf("%v: %s", tt.cmd, "the process has been exited")
		}
		err = KillCommand(cmd)
		if err != nil && !tt.processFinished {
			t.Fatalf("%v: %v", tt.cmd, err)
		}
		if checkprocess() {
			t.Errorf("%v: %s", tt.cmd, "the process has not exited")
		}
	}
}

func TestCommandCancel(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.cmd), func(t *testing.T) {
			_ = killprocess()
			if checkprocess() {
				t.Fatal("the process has not exited")
			}

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			cmd := CommandContext(ctx, tt.cmd[0], tt.cmd[1:]...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Start()
			if err != nil {
				t.Fatalf("%v: %v", tt.cmd, err)
			}
			if !checkprocess() && !tt.processFinished {
				t.Errorf("%v: %s", tt.cmd, "the process has been exited")
			}
			_ = cmd.Cancel()
			if checkprocess() {
				t.Errorf("%v: %s", tt.cmd, "the process has not exited")
			}
		})
	}
}

type testcase struct {
	name            string
	cmd             []string
	want            string
	processFinished bool
}

func checkprocess() bool {
	time.Sleep(500 * time.Millisecond)
	var (
		out []byte
		err error
	)
	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/c", "tasklist | grep stubcmd.exe | grep -v grep").Output()
	} else {
		out, err = exec.Command("bash", "-c", "ps aux | grep stubcmd | grep -v grep").Output()
	}
	return (err == nil || strings.TrimRight(string(out), "\n\r") != "")
}

func killprocess() error {
	var (
		out []byte
		err error
	)
	if runtime.GOOS == "windows" {
		out, err = exec.Command("taskkill", "/im", "stubcmd.exe").Output()
	} else {
		out, err = exec.Command("bash", "-c", "ps aux | grep stubcmd | grep -v grep | xargs kill").Output()
	}
	if err != nil {
		if strings.TrimRight(string(out), "\n\r") != "" {
			_, _ = fmt.Fprintf(os.Stderr, "%s", string(out))
		}
		return err
	}
	return nil
}

func gentests(withSleepTest bool) []testcase {
	tests := []testcase{}
	tests = append(tests, testcase{"echo", []string{stubCmd, "-echo", "!!!"}, "!!!", true})
	tests = append(tests, testcase{"sh -c echo", []string{shellcmd, shellflag, fmt.Sprintf("%s -echo %s", stubCmd, "!!!")}, "!!!", true})
	if withSleepTest {
		tests = append(tests, testcase{"sleep", []string{stubCmd, "-sleep", "10", "-echo", "!!!"}, "!!!", false})
		tests = append(tests, testcase{"sh -c sleep", []string{shellcmd, shellflag, fmt.Sprintf("%s -sleep %s -echo %s", stubCmd, "10", "!!!")}, "!!!", false})
	}
	return tests
}
