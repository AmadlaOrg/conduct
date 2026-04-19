package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result holds the output of a remote command execution.
type Result struct {
	Node     string `json:"node"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// Executor defines the interface for executing commands on remote nodes.
type Executor interface {
	Run(host string, port int, user string, key string, command string) (*Result, error)
	Copy(host string, port int, user string, key string, localPath string, remotePath string) error
}

type executor struct{}

// ExecCommand is a package-level variable for exec.Command, replaceable in tests.
var ExecCommand = exec.Command

// New creates a new executor.
func New() Executor {
	return &executor{}
}

// Run executes a command on a remote node via SSH.
func (e *executor) Run(host string, port int, user string, key string, command string) (*Result, error) {
	args := buildSSHArgs(host, port, user, key)
	args = append(args, command)

	cmd := ExecCommand("ssh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("ssh execution failed: %w", err)
		}
	}

	return &Result{
		Node:     fmt.Sprintf("%s@%s:%d", user, host, port),
		Command:  command,
		ExitCode: exitCode,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
	}, nil
}

// Copy copies a file to a remote node via SCP.
func (e *executor) Copy(host string, port int, user string, key string, localPath string, remotePath string) error {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=" + os.DevNull,
		"-P", fmt.Sprintf("%d", port),
	}

	if key != "" {
		args = append(args, "-i", key)
	}

	dest := fmt.Sprintf("%s@%s:%s", user, host, remotePath)
	args = append(args, localPath, dest)

	cmd := ExecCommand("scp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

func buildSSHArgs(host string, port int, user string, key string) []string {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=" + os.DevNull,
		"-o", "BatchMode=yes",
		"-p", fmt.Sprintf("%d", port),
	}

	if key != "" {
		args = append(args, "-i", key)
	}

	args = append(args, fmt.Sprintf("%s@%s", user, host))
	return args
}
