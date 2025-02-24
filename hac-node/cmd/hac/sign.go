package main

import (
	"encoding/base64"
	"encoding/hex"

	"fmt"

	"github.com/hetu-project/hetu-chaoschain/crypto"

	"github.com/spf13/cobra"
)

type signArguments struct {
	Skey string
}

var signArgs signArguments

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "",
	Long:  ``,
	Run:   signRun,
}

func init() {
	signCmd.Flags().StringVarP(&signArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
}

func signRun(cmd *cobra.Command, args []string) {
	dat := []byte("Hello, CosmJS!")
	pv := crypto.LoadFilePV(signArgs.Skey)
	sig, err := pv.Sign(dat)
	if err != nil {
		fmt.Printf("sign tx err:%v\n", err)
		return
	}
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
	println("signature base64:", base64.StdEncoding.EncodeToString(sig))
	println("signature:", hex.EncodeToString(sig))
}
