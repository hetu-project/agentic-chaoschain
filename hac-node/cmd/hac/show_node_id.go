package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/p2p"
)

var keyfilePath string

func init() {
	showNodeIDCmd.Flags().StringVarP(&keyfilePath, "key", "k", "", "p2p node key path")
}

var showNodeIDCmd = &cobra.Command{
	Use:     "show-node-id",
	Aliases: []string{"show_node_id"},
	Short:   "Show this node's ID",
	RunE:    showNodeID,
}

func showNodeID(*cobra.Command, []string) error {
	nodeKey, err := p2p.LoadNodeKey(keyfilePath)
	if err != nil {
		return err
	}

	fmt.Println(nodeKey.ID())
	return nil
}
