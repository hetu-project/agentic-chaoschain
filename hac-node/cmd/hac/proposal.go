package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/calehh/hac-app/crypto"
	"github.com/calehh/hac-app/tx"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/spf13/cobra"
)

type newProposalArguments struct {
	Url      string
	Index    uint64
	Nonce    uint64
	Skey     string
	Data     string
	NoSend   bool
	Sig      string
	Title    string
	AgentUrl string
}

var newProposalArgs newProposalArguments

var newProposalCmd = &cobra.Command{
	Use:   "newproposal",
	Short: "",
	Long:  ``,
	Run:   newProposalRun,
}

func init() {
	urlFlag(newProposalCmd, &newProposalArgs.Url)
	newProposalCmd.Flags().Uint64VarP(&newProposalArgs.Index, "index", "i", 0, "account index")
	newProposalCmd.Flags().Uint64VarP(&newProposalArgs.Nonce, "nonce", "n", 0, "account nonce")
	newProposalCmd.Flags().StringVarP(&newProposalArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
	newProposalCmd.Flags().StringVarP(&newProposalArgs.Data, "data", "d", "", "proposal data")
	newProposalCmd.Flags().BoolVarP(&newProposalArgs.NoSend, "nosend", "", false, "not send transaction but print signature")
	newProposalCmd.Flags().StringVarP(&newProposalArgs.Sig, "sig", "", "", "transaction signatures")
	newProposalCmd.Flags().StringVarP(&newProposalArgs.Title, "title", "t", "New Proposal", "proposal title")
	newProposalCmd.Flags().StringVarP(&newProposerArgs.AgentUrl, "agent", "a", "http://127.0.0.1/3000", "agent")
}

func newProposalRun(cmd *cobra.Command, args []string) {
	cli, err := http.New(newProposalArgs.Url, "/websocket")
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
	nonce := newProposalArgs.Nonce
	if nonce == 0 {
		act, err := queryAccount(newProposalArgs.Url, newProposalArgs.Index, "")
		if err != nil {
			return
		}
		nonce = act.Nonce
	}
	btx := tx.HACTx{
		Version:   tx.HACTxVersion1,
		Nonce:     nonce,
		Validator: newProposalArgs.Index,
	}
	if newProposalArgs.Title == "" {
		newProposalArgs.Title, err = summarizePR(PR{Data: newProposalArgs.Data})
		if err != nil {
			fmt.Printf("summarize proposal err:%v\n", err)
			return
		}
	}
	stx := &tx.ProposalTx{
		Proposer:  newProposalArgs.Index,
		EndHeight: 1000000,
		ImageUrl:  "",
		Title:     newProposalArgs.Title,
		Link:      "",
		Data:      []byte(newProposalArgs.Data),
	}
	btx.Tx = stx
	btx.Type = tx.HACTxTypeProposal
	dat, err := btx.SigData([]byte(chainId))
	if err != nil {
		fmt.Printf("tx sign data err:%v\n", err)
		return
	}
	println("data signed:", hex.EncodeToString(dat))
	sigs := [][]byte{}
	pv := crypto.LoadFilePV(newProposalArgs.Skey)
	sig, err := pv.Sign(dat)
	if err != nil {
		fmt.Printf("sign tx err:%v\n", err)
		return
	}
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
	sigs = append(sigs, sig)
	if newProposalArgs.NoSend {
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
