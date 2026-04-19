package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AmadlaOrg/conduct/executor"
	"github.com/AmadlaOrg/conduct/state"
	"github.com/spf13/cobra"
)

var (
	execStateNew = state.New

	// ExecCmd runs a command on a specific node.
	ExecCmd = &cobra.Command{
		Use:   "exec <deployment> <node> -- <command...>",
		Short: "Execute a command on a specific node",
		Long:  "Runs an arbitrary command on a node within a deployment via SSH.",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runExec,
	}
)

func runExec(cmd *cobra.Command, args []string) error {
	deploymentName := args[0]
	nodeName := args[1]

	// Find everything after "--"
	var remoteCmd string
	dashDash := false
	for _, a := range os.Args {
		if a == "--" {
			dashDash = true
			continue
		}
		if dashDash {
			if remoteCmd != "" {
				remoteCmd += " "
			}
			remoteCmd += a
		}
	}

	if remoteCmd == "" && len(args) > 2 {
		remoteCmd = strings.Join(args[2:], " ")
	}

	if remoteCmd == "" {
		return fmt.Errorf("no command specified (use -- before the command)")
	}

	stateMgr := execStateNew()
	ds, err := stateMgr.Load(deploymentName)
	if err != nil {
		return fmt.Errorf("deployment %q not found: %w", deploymentName, err)
	}

	var nodeState *state.NodeState
	for i := range ds.Nodes {
		if ds.Nodes[i].Name == nodeName {
			nodeState = &ds.Nodes[i]
			break
		}
	}

	if nodeState == nil {
		return fmt.Errorf("node %q not found in deployment %q", nodeName, deploymentName)
	}

	exec := executor.New()
	result, err := exec.Run(nodeState.Host, 22, "root", "", remoteCmd)
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	if result.Stdout != "" {
		fmt.Println(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprintln(os.Stderr, result.Stderr)
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}
