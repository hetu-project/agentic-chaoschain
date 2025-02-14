package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/calehh/hac-app/crypto"
	"github.com/calehh/hac-app/tx"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/spf13/cobra"
)

type settleArguments struct {
	Url      string
	Index    uint64
	Nonce    uint64
	Skey     string
	Proposal uint64
	NoSend   bool
	Sig      string
}

var settleArgs settleArguments

var settleCmd = &cobra.Command{
	Use:   "settle",
	Short: "",
	Long:  ``,
	Run:   settleRun,
}

func init() {
	urlFlag(settleCmd, &settleArgs.Url)
	settleCmd.Flags().Uint64VarP(&settleArgs.Index, "index", "i", 0, "account index")
	settleCmd.Flags().Uint64VarP(&settleArgs.Nonce, "nonce", "n", 0, "account nonce")
	settleCmd.Flags().StringVarP(&settleArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
	settleCmd.Flags().Uint64VarP(&settleArgs.Proposal, "proposal", "p", 0, "proposal index")
	settleCmd.Flags().BoolVarP(&settleArgs.NoSend, "nosend", "", false, "not send transaction but print signature")
	settleCmd.Flags().StringVarP(&settleArgs.Sig, "sig", "", "", "transaction signatures")
}

func settleRun(cmd *cobra.Command, args []string) {
	cli, err := http.New(settleArgs.Url, "/websocket")
	if err != nil {
		fmt.Printf("new client err:%v\n", err)
		return
	}
	ctx := context.Background()
	gres, err := cli.Genesis(ctx)
	if err != nil {
		fmt.Printf("get chain genesis err:%v\n", err)
		return
	}
	chainId := gres.Genesis.ChainID
	nonce := settleArgs.Nonce
	if nonce == 0 {
		act, err := queryAccount(settleArgs.Url, settleArgs.Index, "")
		if err != nil {
			return
		}
		nonce = act.Nonce
	}
	btx := tx.HACTx{
		Version:   tx.HACTxVersion1,
		Nonce:     nonce,
		Validator: settleArgs.Index,
	}
	stx := &tx.SettleProposalTx{
		Proposal:        settleArgs.Proposal,
		ExpireTimestamp: uint(time.Now().Unix() + 60*3),
	}
	btx.Tx = stx
	btx.Type = tx.HACTxTypeSettleProposal
	dat, err := btx.SigData([]byte(chainId))
	if err != nil {
		fmt.Printf("tx sign data err:%v\n", err)
		return
	}
	println("data signed:", hex.EncodeToString(dat))
	sigs := [][]byte{}
	pv := crypto.LoadFilePV(settleArgs.Skey)
	sig, err := pv.Sign(dat)
	if err != nil {
		fmt.Printf("sign tx err:%v\n", err)
		return
	}
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
	sigs = append(sigs, sig)
	if settleArgs.NoSend {
		fmt.Println("transaction signatures:")
		for _, sig := range sigs {
			fmt.Println(hex.EncodeToString(sig))
		}
		return
	}
	btx.Sig = sigs
	dat, err = json.Marshal(btx)
	if err != nil {
		fmt.Printf("rlp encode tx err:%v\n", err)
		return
	}
	fmt.Printf("tx:%x btx:%#v\n", dat, btx)
	res, err := cli.BroadcastTxSync(ctx, dat)
	if err != nil {
		fmt.Printf("broadcast tx err:%v\n", err)
		return
	}
	dat, _ = json.Marshal(res)
	fmt.Printf("%v\n", string(dat))
}
