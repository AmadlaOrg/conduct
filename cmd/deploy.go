package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AmadlaOrg/conduct/executor"
	"github.com/AmadlaOrg/conduct/plan"
	"github.com/AmadlaOrg/conduct/state"
	"github.com/AmadlaOrg/conduct/topology"
	"github.com/spf13/cobra"
)

var (
	deployFilePath string
	deployDryRun   bool

	// DeployCmd executes a multi-node deployment.
	DeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy infrastructure across multiple nodes",
		Long:  "Reads a topology file and executes tools on each node in dependency order.",
		RunE:  runDeploy,
	}
)

func init() {
	DeployCmd.Flags().StringVarP(&deployFilePath, "file", "f", "", "Topology file (JSON or YAML)")
	_ = DeployCmd.MarkFlagRequired("file")
	DeployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Print execution plan without running")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	var input io.Reader

	if deployFilePath == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(deployFilePath)
		if err != nil {
			return fmt.Errorf("cannot open file: %w", err)
		}
		defer f.Close()
		input = f
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("cannot read input: %w", err)
	}

	topo, err := topology.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse topology: %w", err)
	}

	execPlan, err := plan.Build(topo)
	if err != nil {
		return fmt.Errorf("failed to build execution plan: %w", err)
	}

	if deployDryRun {
		fmt.Fprintf(os.Stderr, "Deployment plan for %q (%d steps):\n\n", execPlan.Name, len(execPlan.Steps))
		for _, step := range execPlan.Steps {
			fmt.Fprintf(os.Stderr, "  %d. [%s] %s@%s:%d -> %s\n",
				step.Order, step.Node, step.User, step.Host, step.Port, step.Tool)
			for k, v := range step.Vars {
				fmt.Fprintf(os.Stderr, "     var %s=%s\n", k, v)
			}
		}
		return nil
	}

	// Execute the plan
	exec := executor.New()
	stateMgr := state.New()

	now := time.Now().UTC().Format(time.RFC3339)
	deployState := &state.DeploymentState{
		Name:      topo.Name,
		Status:    "deploying",
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, node := range topo.Nodes {
		deployState.Nodes = append(deployState.Nodes, state.NodeState{
			Name:   node.Name,
			Host:   node.Host,
			Status: "pending",
		})
	}

	_ = stateMgr.Save(deployState)

	for _, step := range execPlan.Steps {
		fmt.Fprintf(os.Stderr, "[%d/%d] %s: running %s on %s...\n",
			step.Order, len(execPlan.Steps), step.Node, step.Tool, step.Host)

		// Build the remote command
		remoteCmd := step.Tool
		for _, arg := range step.Args {
			remoteCmd += " " + arg
		}

		// Set environment variables for cross-node data
		for k, v := range step.Vars {
			remoteCmd = fmt.Sprintf("export %s=%q; %s", k, v, remoteCmd)
		}

		result, err := exec.Run(step.Host, step.Port, step.User, step.Key, remoteCmd)
		if err != nil {
			updateNodeStatus(deployState, step.Node, "failed")
			deployState.Status = "failed"
			deployState.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			_ = stateMgr.Save(deployState)
			return fmt.Errorf("step %d failed on %s: %w", step.Order, step.Node, err)
		}

		if result.ExitCode != 0 {
			updateNodeStatus(deployState, step.Node, "failed")
			deployState.Status = "failed"
			deployState.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			_ = stateMgr.Save(deployState)
			return fmt.Errorf("step %d failed on %s (exit %d): %s",
				step.Order, step.Node, result.ExitCode, result.Stderr)
		}

		if result.Stdout != "" {
			fmt.Println(result.Stdout)
		}

		updateNodeStatus(deployState, step.Node, "ready")
	}

	deployState.Status = "deployed"
	deployState.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_ = stateMgr.Save(deployState)

	fmt.Fprintf(os.Stderr, "\nDeployment %q complete (%d steps)\n", topo.Name, len(execPlan.Steps))
	return outputResult(deployState)
}

func updateNodeStatus(ds *state.DeploymentState, nodeName, status string) {
	for i := range ds.Nodes {
		if ds.Nodes[i].Name == nodeName {
			ds.Nodes[i].Status = status
			return
		}
	}
}
