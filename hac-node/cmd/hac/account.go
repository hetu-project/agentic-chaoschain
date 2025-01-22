package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/calehh/hac-app/state"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	DefaultPrivValKeyName   = "priv_validator_key.json"
	DefaultPrivValStateName = "priv_validator_state.json"
)

type accountArguments struct {
	Url     string
	Address string
	Index   uint64
}

var accountArgs accountArguments

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "",
	Long:  ``,
	Run:   accountRun,
}

func init() {
	urlFlag(accountCmd, &accountArgs.Url)
	accountCmd.Flags().StringVarP(&accountArgs.Address, "address", "a", "", "account address")
	accountCmd.Flags().Uint64VarP(&accountArgs.Index, "index", "i", 0, "account index")
	showCmd.Flags().StringVarP(&showArgs.Home, "homedir", "d", "data", "home dir")
	accountCmd.AddCommand(showCmd)
}

func accountRun(cmd *cobra.Command, args []string) {
	act, err := queryAccount(accountArgs.Url, accountArgs.Index, accountArgs.Address)
	if err != nil {
		return
	}
	pk := ed25519.PubKey(act.PubKey[:])
	actStr := fmt.Sprintf("name:%s nonce:%v index:%v pk:%v stake:%v addr:%v agentUrl:%s\n",
		act.Name, act.Nonce, act.Index, common.Bytes2Hex(act.PubKey), act.Stake, common.Bytes2Hex(pk.Address()[:]), act.AgentUrl)
	fmt.Println(actStr)
}

func queryAccount(url string, index uint64, address string) (*state.Account, error) {
	cli, err := http.New(url, "/websocket")
	if err != nil {
		fmt.Printf("new client err:%v\n", err)
		return nil, err
	}
	ctx := context.Background()
	var dat []byte
	if len(address) > 0 {
		dat, err = hex.DecodeString(address)
		if err != nil {
			fmt.Printf("invalid address:%v\n", accountArgs.Address)
			return nil, err
		}
	} else {
		s := fmt.Sprintf("0%x", index)
		if len(s)&1 == 1 {
			s = s[1:]
		}
		dat, _ = hex.DecodeString(s)
	}
	res, err := cli.ABCIQuery(ctx, "/accounts/", dat)
	if err != nil {
		fmt.Printf("request err:%v\n", err)
		return nil, err
	}
	if res.Response.Code != 0 {
		fmt.Printf("%#v\n", res)
		return nil, errors.New("response code 0")
	}
	var act state.Account
	err = act.UnmarshalJSON(res.Response.Value)
	if err != nil {
		return nil, err
	}
	return &act, err
}

type showArguments struct {
	Home string
}

var showArgs showArguments

var showCmd = &cobra.Command{
	Use:   "pk",
	Short: "",
	Long:  ``,
	Run:   showRun,
}

func showRun(cmd *cobra.Command, args []string) {
	filePV := privval.LoadFilePV(
		filepath.Join(showArgs.Home, "config", DefaultPrivValKeyName),
		filepath.Join(showArgs.Home, "data", DefaultPrivValStateName),
	)
	pubKey, err := filePV.GetPubKey()
	if err != nil {
		fmt.Printf("get public key error %v", err)
		return
	}
	fmt.Printf("pk:%s\n", hex.EncodeToString(pubKey.Bytes()))
}
