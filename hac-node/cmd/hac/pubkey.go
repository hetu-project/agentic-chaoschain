package main

import (
	"encoding/hex"

	"github.com/hetu-project/hetu-chaoschain/crypto"
	"github.com/spf13/cobra"
)

type pubkeyArguments struct {
	Skey string
}

var pubkeyArgs pubkeyArguments

var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "",
	Long:  ``,
	Run:   pubkeyRun,
}

func init() {
	pubkeyCmd.Flags().StringVarP(&pubkeyArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
}

func pubkeyRun(cmd *cobra.Command, args []string) {
	pv := crypto.LoadFilePV(pubkeyArgs.Skey)
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
}
