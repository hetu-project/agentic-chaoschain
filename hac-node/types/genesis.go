package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
)

type GenesisState map[string]json.RawMessage

type GenesisValidator struct {
	Address crypto.Address `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	Power   int64          `json:"power"`
	Name    string         `json:"name"`
}

type AgentInfo struct {
	Address  string `json:"address"`
	AgentUrl string `json:"agent_url"`
	Name     string `json:"name"`
}

type GenesisAppState struct {
	Agents   []AgentInfo `json:"agents"`
	Manifest string      `json:"manifest"`
}

// GenesisDoc defines the initial conditions for a CometBFT blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time                 `json:"genesis_time"`
	ChainID         string                    `json:"chain_id"`
	InitialHeight   int64                     `json:"initial_height"`
	ConsensusParams *cmttypes.ConsensusParams `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator        `json:"validators"`
	AppHash         []byte                    `json:"app_hash"`
	AppState        json.RawMessage           `json:"app_state"`
}

// SaveAs is a utility method for saving GenensisDoc as a JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes, err := cmtjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, genDocBytes, 0o600)
}

func (ag *GenesisDoc) ValidateAndComplete() error {
	if ag.ChainID == "" {
		return errors.New("genesis doc must include non-empty chain_id")
	}

	if ag.InitialHeight < 0 {
		return fmt.Errorf("initial_height cannot be negative (got %v)", ag.InitialHeight)
	}

	if ag.InitialHeight == 0 {
		ag.InitialHeight = 1
	}

	if ag.GenesisTime.IsZero() {
		ag.GenesisTime = time.Now().Round(0).UTC()
	}

	return nil
}

func ExportGenesisFile(genesis *GenesisDoc, genFile string) error {
	if err := genesis.ValidateAndComplete(); err != nil {
		return err
	}
	return genesis.SaveAs(genFile)
}

const HACModuleName = "hac"
const DefaultPower = 1000
const DefaultStatement = "Learn! Build! Thrive!"
const DefaultAgentUrl = "http://127.0.0.1:3000"
