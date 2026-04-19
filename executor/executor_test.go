package executor

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeExecCommand(exitCode int, stdout string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			fmt.Sprintf("GO_HELPER_EXIT_CODE=%d", exitCode),
			fmt.Sprintf("GO_HELPER_STDOUT=%s", stdout),
		)
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	exitCode := 0
	if code := os.Getenv("GO_HELPER_EXIT_CODE"); code != "" {
		fmt.Sscanf(code, "%d", &exitCode)
	}
	stdout := os.Getenv("GO_HELPER_STDOUT")
	if stdout != "" {
		fmt.Fprint(os.Stdout, stdout)
	}
	os.Exit(exitCode)
}

func TestExecutor_Run_Success(t *testing.T) {
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()
	ExecCommand = fakeExecCommand(0, "hello world")

	ex := New()
	result, err := ex.Run("10.0.0.1", 22, "root", "", "echo hello")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello world", result.Stdout)
	assert.Equal(t, "echo hello", result.Command)
}

func TestExecutor_Run_NonZeroExit(t *testing.T) {
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()
	ExecCommand = fakeExecCommand(1, "")

	ex := New()
	result, err := ex.Run("10.0.0.1", 22, "root", "", "false")
	require.NoError(t, err)
	assert.Equal(t, 1, result.ExitCode)
}

func TestExecutor_Run_WithKey(t *testing.T) {
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()
	ExecCommand = fakeExecCommand(0, "ok")

	ex := New()
	result, err := ex.Run("10.0.0.1", 2222, "ubuntu", "~/.ssh/id_rsa", "whoami")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Node, "ubuntu@10.0.0.1:2222")
}

func TestExecutor_Copy_Success(t *testing.T) {
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()
	ExecCommand = fakeExecCommand(0, "")

	ex := New()
	err := ex.Copy("10.0.0.1", 22, "root", "", "/tmp/local.txt", "/tmp/remote.txt")
	assert.NoError(t, err)
}

func TestExecutor_Copy_Failure(t *testing.T) {
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()
	ExecCommand = fakeExecCommand(1, "scp error")

	ex := New()
	err := ex.Copy("10.0.0.1", 22, "root", "", "/tmp/local.txt", "/tmp/remote.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scp failed")
}

func TestBuildSSHArgs_Default(t *testing.T) {
	args := buildSSHArgs("10.0.0.1", 22, "root", "")
	assert.Contains(t, args, "-p")
	assert.Contains(t, args, "22")
	assert.Contains(t, args, "root@10.0.0.1")
	assert.NotContains(t, args, "-i")
}

func TestBuildSSHArgs_WithKey(t *testing.T) {
	args := buildSSHArgs("10.0.0.1", 2222, "ubuntu", "/home/user/.ssh/key")
	assert.Contains(t, args, "-i")
	assert.Contains(t, args, "/home/user/.ssh/key")
	assert.Contains(t, args, "2222")
	assert.Contains(t, args, "ubuntu@10.0.0.1")
}
