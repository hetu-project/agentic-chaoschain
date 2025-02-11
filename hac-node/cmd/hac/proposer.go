package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	htp "net/http"

	"github.com/calehh/hac-app/crypto"
	"github.com/calehh/hac-app/tx"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/spf13/cobra"
)

type newProposerArguments struct {
	Url       string
	Index     uint64
	Nonce     uint64
	Skey      string
	NoSend    bool
	Sig       string
	SourceUrl string
	Duration  uint64
	AgentUrl  string
}

var newProposerArgs newProposerArguments

var newProposerCmd = &cobra.Command{
	Use:   "newProposer",
	Short: "",
	Long:  ``,
	Run:   newProposerRun,
}

func init() {
	urlFlag(newProposerCmd, &newProposerArgs.Url)
	newProposerCmd.Flags().Uint64VarP(&newProposerArgs.Index, "index", "i", 0, "account index")
	newProposerCmd.Flags().Uint64VarP(&newProposerArgs.Nonce, "nonce", "n", 0, "account nonce")
	newProposerCmd.Flags().StringVarP(&newProposerArgs.Skey, "skeyPath", "s", "./config/priv_validator_key.json", "private key path")
	newProposerCmd.Flags().BoolVarP(&newProposerArgs.NoSend, "nosend", "", false, "not send transaction but print signature")
	newProposerCmd.Flags().StringVarP(&newProposerArgs.Sig, "sig", "", "", "transaction signatures")
	newProposerCmd.Flags().StringVarP(&newProposerArgs.SourceUrl, "source", "o", "", "source url")
	newProposerCmd.Flags().Uint64VarP(&newProposerArgs.Duration, "duration", "t", 60, "duration")
	newProposerCmd.Flags().StringVarP(&newProposerArgs.AgentUrl, "agent", "a", "http://127.0.0.1/3000", "agent")
}

type PR struct {
	Title    string `json:"title"`
	Data     string `json:"data"`
	Link     string `json:"link"`
	ImageUrl string `json:"imageUrl"`
}

func getLatestPR() (PR, error) {
	return PR{}, nil
}

type SummarizeResp struct {
	Title string `json:"title"`
}

func summarizePR(pr PR) (string, error) {
	// post request agent url
	data := map[string]interface{}{
		"text": pr.Data,
	}
	d, _ := json.Marshal(data)
	resp, err := htp.Post(newProposerArgs.AgentUrl, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var sr SummarizeResp
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return "", err
	}
	return sr.Title, nil
}

func newProposerRun(cmd *cobra.Command, args []string) {
	ticker := time.NewTicker(time.Duration(newProposerArgs.Duration) * time.Minute)
	for {
		<-ticker.C
		cli, err := http.New(newProposerArgs.Url, "/websocket")
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
		nonce := newProposerArgs.Nonce
		if nonce == 0 {
			act, err := queryAccount(newProposerArgs.Url, newProposerArgs.Index, "")
			if err != nil {
				return
			}
			nonce = act.Nonce
		}
		btx := tx.HACTx{
			Version:   tx.HACTxVersion1,
			Nonce:     nonce,
			Validator: newProposerArgs.Index,
		}
		pr, err := getLatestPR()
		if err != nil {
			fmt.Printf("get latest pr err:%v\n", err)
			continue
		}
		title, err := summarizePR(pr)
		if err != nil {
			fmt.Printf("summarize pr err:%v\n", err)
			continue
		}
		stx := &tx.ProposalTx{
			Proposer:  newProposerArgs.Index,
			EndHeight: 0,
			ImageUrl:  pr.ImageUrl,
			Title:     title,
			Link:      pr.Link,
			Data:      []byte(pr.Data),
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
		pv := crypto.LoadFilePV(newProposerArgs.Skey)
		sig, err := pv.Sign(dat)
		if err != nil {
			fmt.Printf("sign tx err:%v\n", err)
			return
		}
		println("pubkey:", hex.EncodeToString(pv.PublicKey()))
		println("address:", pv.Address())
		sigs = append(sigs, sig)
		if newProposerArgs.NoSend {
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
}
