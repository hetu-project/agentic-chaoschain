package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/hetu-project/hetu-chaoschain/crypto"
	"github.com/hetu-project/hetu-chaoschain/tx"
	"github.com/spf13/cobra"
)

type discussionArguments struct {
	Url      string
	Index    uint64
	Nonce    uint64
	Skey     string
	Data     string
	Proposal uint64
	NoSend   bool
	Sig      string
}

var discussionArgs discussionArguments

var discussionCmd = &cobra.Command{
	Use:   "discussion",
	Short: "",
	Long:  ``,
	Run:   discussionRun,
}

func init() {
	urlFlag(discussionCmd, &discussionArgs.Url)
	discussionCmd.Flags().Uint64VarP(&discussionArgs.Index, "index", "i", 0, "account index")
	discussionCmd.Flags().Uint64VarP(&discussionArgs.Nonce, "nonce", "n", 0, "account nonce")
	discussionCmd.Flags().StringVarP(&discussionArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
	discussionCmd.Flags().StringVarP(&discussionArgs.Data, "data", "d", "", "proposal data")
	discussionCmd.Flags().Uint64VarP(&discussionArgs.Proposal, "proposal", "p", 0, "proposal index")
	discussionCmd.Flags().BoolVarP(&discussionArgs.NoSend, "nosend", "", false, "not send transaction but print signature")
	discussionCmd.Flags().StringVarP(&discussionArgs.Sig, "sig", "", "", "transaction signatures")
}

func discussionRun(cmd *cobra.Command, args []string) {
	cli, err := http.New(discussionArgs.Url, "/websocket")
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
	nonce := discussionArgs.Nonce
	if nonce == 0 {
		act, err := queryAccount(discussionArgs.Url, discussionArgs.Index, "")
		if err != nil {
			return
		}
		nonce = act.Nonce
	}
	btx := tx.HACTx{
		Version:   tx.HACTxVersion1,
		Nonce:     nonce,
		Validator: discussionArgs.Index,
	}
	stx := &tx.DiscussionTx{
		Proposal: discussionArgs.Proposal,
		Data:     []byte(discussionArgs.Data),
	}
	btx.Tx = stx
	btx.Type = tx.HACTxTypeDiscussion
	dat, err := btx.SigData([]byte(chainId))
	if err != nil {
		fmt.Printf("tx sign data err:%v\n", err)
		return
	}
	println("data to sign:", string(dat))
	println("data signed:", hex.EncodeToString(dat))
	sigs := [][]byte{}
	pv := crypto.LoadFilePV(discussionArgs.Skey)
	sig, err := pv.Sign(dat)
	if err != nil {
		fmt.Printf("sign tx err:%v\n", err)
		return
	}
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
	sigs = append(sigs, sig)
	if discussionArgs.NoSend {
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
