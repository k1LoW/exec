package exec

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	stubCmd = `.\testdata\stubcmd.exe`
)

func init() {
	if runtime.GOOS != "windows" {
		return
	}
	err := exec.Command("go", "build", "-o", stubCmd, "testdata/stubcmd.go").Run()
	if err != nil {
		panic(err)
	}
}

func TestCommand(t *testing.T) {
	tests := gentests(false)
	for _, tt := range tests {
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if strings.TrimRight(stdout.String(), "\n\r") != tt.want {
			t.Errorf("%s: want = %#v, got = %#v", tt.name, tt.want, stdout.String())
		}
		_, err = exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep", tt.want)).Output()
		if err == nil {
			t.Errorf("%s", "the process has not exited")
		}
	}
}

func TestCommandContext(t *testing.T) {
	tests := gentests(false)
	for _, tt := range tests {
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
			t.Fatalf("%s: %v", tt.name, err)
		}
		if strings.TrimRight(stdout.String(), "\n\r") != tt.want {
			t.Errorf("%s: want = %#v, got = %#v", tt.name, tt.want, stdout.String())
		}
		_, err = exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep", tt.want)).Output()
		if err == nil {
			t.Errorf("%s", "the process has not exited")
		}
	}
}

func TestCommandContextCancel(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cmd := CommandContext(ctx, tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		go func() {
			cmd.Wait()
		}()
		time.Sleep(100 * time.Millisecond)
		cancel()
		_, err = exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep", tt.want)).Output()
		if err == nil {
			t.Errorf("%s", "the process has not exited")
		}
	}
}

func TestTerminateCommand(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			t.Fatalf("%v", err)
		}
		go func() {
			cmd.Wait()
		}()
		time.Sleep(500 * time.Millisecond)
		if runtime.GOOS == "windows" {
			sig := os.Interrupt
			err = TerminateCommand(cmd, sig)
		} else {
			sig := syscall.SIGTERM
			err = TerminateCommand(cmd, sig)
		}
		if err != nil && !tt.processFinished {
			t.Errorf("%s: %v", tt.name, err)
		}
		_, err = exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep", tt.want)).Output()
		if err == nil {
			t.Errorf("%s: %s", tt.name, "the process has not exited")
		}
	}
}

func TestKillCommand(t *testing.T) {
	tests := gentests(true)
	for _, tt := range tests {
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		cmd := Command(tt.cmd[0], tt.cmd[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			t.Fatalf("%v", err)
		}
		go func() {
			cmd.Wait()
		}()
		time.Sleep(500 * time.Millisecond)
		err = KillCommand(cmd)
		if err != nil && !tt.processFinished {
			t.Fatalf("%s: %v", tt.name, err)
		}
		_, err = exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep", tt.want)).Output()
		if err == nil {
			t.Errorf("%s: %s", tt.name, "the process has not exited")
		}
	}
}

type testcase struct {
	name            string
	cmd             []string
	want            string
	processFinished bool
}

func gentests(withSleepTest bool) []testcase {
	tests := []testcase{}
	if runtime.GOOS == "windows" {
		r := random()
		tests = append(tests, testcase{"echo", []string{"echo", r}, r, true})
		r = random()
		tests = append(tests, testcase{"cmd /c echo", []string{"cmd", "/c", fmt.Sprintf("echo %s", r)}, r, true})
		if withSleepTest {
			r = "123456"
			tests = append(tests, testcase{"sleep", []string{stubCmd, "-sleep", r}, r, false})
			r = "654321"
			tests = append(tests, testcase{"cmd /c sleep", []string{"cmd", "/c", fmt.Sprintf("%s -sleep %s && echo %s", stubCmd, r, r)}, r, false})
		}
		return tests
	}
	r := random()
	tests = append(tests, testcase{"echo", []string{"echo", r}, r, true})
	r = random()
	tests = append(tests, testcase{"bash -c echo", []string{"bash", "-c", fmt.Sprintf("echo %s", r)}, r, true})
	if withSleepTest {
		r = "123456"
		tests = append(tests, testcase{"sleep", []string{"sleep", r}, r, false})
		r = "654321"
		tests = append(tests, testcase{"bash -c sleep", []string{"bash", "-c", fmt.Sprintf("sleep %s && echo %s", r, r)}, r, false})
	}
	return tests
}

func random() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(rand.Int())
}
