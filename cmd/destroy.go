package cmd

import (
	"fmt"

	"github.com/AmadlaOrg/conduct/state"
	"github.com/spf13/cobra"
)

var (
	destroyStateNew = state.New

	// DestroyCmd removes a deployment.
	DestroyCmd = &cobra.Command{
		Use:   "destroy <deployment>",
		Short: "Remove a deployment record",
		Args:  cobra.ExactArgs(1),
		RunE:  runDestroy,
	}
)

func runDestroy(cmd *cobra.Command, args []string) error {
	stateMgr := destroyStateNew()

	if err := stateMgr.Remove(args[0]); err != nil {
		return fmt.Errorf("failed to remove deployment: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Deployment %q removed\n", args[0])
	return nil
}
