package main

import (
	"fmt"
	"os"
)

func main() {
	clCmd.AddCommand(accountCmd)
	clCmd.AddCommand(initCmd)
	clCmd.AddCommand(versionCmd)
	clCmd.AddCommand(newProposalCmd)
	clCmd.AddCommand(discussionCmd)
	clCmd.AddCommand(settleCmd)
	clCmd.AddCommand(grantCmd)
	clCmd.AddCommand(pubkeyCmd)
	clCmd.AddCommand(signCmd)
	clCmd.AddCommand(mockCmd)
	clCmd.AddCommand(showNodeIDCmd)
	if err := clCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
