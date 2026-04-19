package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/conduct/state"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	statusStateNew = state.New

	// StatusCmd shows deployment status.
	StatusCmd = &cobra.Command{
		Use:   "status [deployment]",
		Short: "Show deployment status",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runStatus,
	}
)

func runStatus(cmd *cobra.Command, args []string) error {
	stateMgr := statusStateNew()

	if len(args) == 1 {
		ds, err := stateMgr.Load(args[0])
		if err != nil {
			return fmt.Errorf("deployment %q not found: %w", args[0], err)
		}

		fmt.Printf("Deployment: %s\n", ds.Name)
		fmt.Printf("Status: %s\n", ds.Status)
		fmt.Printf("Created: %s\n", ds.CreatedAt)
		fmt.Printf("Updated: %s\n\n", ds.UpdatedAt)

		if len(ds.Nodes) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.Header("Node", "Host", "Status")
			for _, n := range ds.Nodes {
				table.Append(n.Name, n.Host, n.Status)
			}
			table.Render()
		}

		return nil
	}

	states, err := stateMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(states) == 0 {
		fmt.Fprintln(os.Stderr, "No deployments found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Deployment", "Status", "Nodes", "Updated")

	for _, s := range states {
		table.Append(s.Name, s.Status, fmt.Sprintf("%d", len(s.Nodes)), s.UpdatedAt)
	}

	table.Render()
	return nil
}
