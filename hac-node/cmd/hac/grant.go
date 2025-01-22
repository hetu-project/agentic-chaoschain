package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/calehh/hac-app/crypto"
	"github.com/calehh/hac-app/tx"
	"github.com/calehh/hac-app/types"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/spf13/cobra"
)

type grantArguments struct {
	Url       string
	Index     uint64
	Nonce     uint64
	Skey      string
	Amount    uint64
	Pubkey    string
	Name      string
	AgentUrl  string
	Statement string
	NoSend    bool
	Sig       string
}

var grantArgs grantArguments

var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "",
	Long:  ``,
	Run:   grantRun,
}

func init() {
	urlFlag(grantCmd, &grantArgs.Url)
	grantCmd.Flags().Uint64VarP(&grantArgs.Index, "index", "i", 0, "account index")
	grantCmd.Flags().Uint64VarP(&grantArgs.Nonce, "nonce", "n", 0, "account nonce")
	grantCmd.Flags().StringVarP(&grantArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
	grantCmd.Flags().StringVarP(&grantArgs.Statement, "statement", "", "", "grant statement")
	grantCmd.Flags().StringVarP(&grantArgs.Pubkey, "pubkey", "p", "", "new account pubkey")
	grantCmd.Flags().Uint64VarP(&grantArgs.Amount, "amount", "a", 0, "grant amout")
	grantCmd.Flags().BoolVarP(&grantArgs.NoSend, "nosend", "", false, "not send transaction but print signature")
	grantCmd.Flags().StringVarP(&grantArgs.Sig, "sig", "", "", "transaction signatures")
	grantCmd.Flags().StringVarP(&grantArgs.Name, "name", "", "", "account name")
	grantCmd.Flags().StringVarP(&grantArgs.AgentUrl, "agentUrl", "", "", "account agentUrl")
}

func grantRun(cmd *cobra.Command, args []string) {
	cli, err := http.New(grantArgs.Url, "/websocket")
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
	nonce := grantArgs.Nonce
	if nonce == 0 {
		act, err := queryAccount(grantArgs.Url, grantArgs.Index, "")
		if err != nil {
			return
		}
		nonce = act.Nonce
	}
	btx := tx.HACTx{
		Version:   tx.HACTxVersion1,
		Nonce:     nonce,
		Validator: grantArgs.Index,
	}
	pubkey, err := hex.DecodeString(grantArgs.Pubkey)
	if err != nil {
		fmt.Printf("decode pubkey hex err:%v\n", err)
		return
	}
	if grantArgs.Name == "" {
		grantArgs.Name = hex.EncodeToString(pubkey)
	}
	if grantArgs.AgentUrl == "" {
		grantArgs.AgentUrl = types.DefaultAgentUrl
	}
	stx := &tx.GrantTx{
		Grants: []tx.GrantSt{
			tx.GrantSt{
				Statement: grantArgs.Statement,
				Amount:    grantArgs.Amount,
				AgentUrl:  grantArgs.AgentUrl,
				Name:      grantArgs.Name,
				Pubkey:    pubkey,
			},
		},
	}
	btx.Tx = stx
	btx.Type = tx.HACTxTypeGrant
	dat, err := btx.SigData([]byte(chainId))
	if err != nil {
		fmt.Printf("tx sign data err:%v\n", err)
		return
	}
	println("data signed:", hex.EncodeToString(dat))
	sigs := [][]byte{}
	pv := crypto.LoadFilePV(grantArgs.Skey)
	sig, err := pv.Sign(dat)
	if err != nil {
		fmt.Printf("sign tx err:%v\n", err)
		return
	}
	println("pubkey:", hex.EncodeToString(pv.PublicKey()))
	println("address:", pv.Address())
	sigs = append(sigs, sig)
	if grantArgs.NoSend {
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
