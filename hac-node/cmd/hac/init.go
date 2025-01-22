package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/calehh/hac-app/config"
	app_config "github.com/calehh/hac-app/config"
	"github.com/calehh/hac-app/types"
	"github.com/cometbft/cometbft/crypto"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/spf13/cobra"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", out)

	return err
}

var initCmd = &cobra.Command{
	Use:   "init hac",
	Short: "Initialize private validator, p2p, genesis, and application configuration files",
	Long:  `Initialize validators's and node's configuration files.`,
	Args:  cobra.ExactArgs(0),
	RunE:  initRun,
}

func init() {
	initCmd.Flags().BoolP(types.FlagOverwrite, "o", false, "overwrite the genesis.json file")
	initCmd.Flags().String(types.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	initCmd.Flags().String(types.FlagHome, "", "config")
}

func initRun(cmd *cobra.Command, args []string) error {
	home, _ := cmd.Flags().GetString(types.FlagHome)
	chainID, _ := cmd.Flags().GetString(types.FlagChainID)
	var (
		genesisTime time.Time
		pk          crypto.PubKey
	)

	switch {
	case chainID != "":
	default:
		chainID = fmt.Sprintf("test-chain-%v", rand.Uint64())
	}
	vals := make([]types.GenesisValidator, 0)
	appConfig := app_config.NewHACConfig(home)

	if chainID != "" {
		chainID = fmt.Sprintf("test-chain-%v", rand.Uint64())
	}
	genesisTime = time.Now()
	_, pk1, err := config.InitializeNodeValidatorFiles(appConfig, nil)
	if err != nil {
		return err
	}
	pk = pk1
	vals = append(vals, types.GenesisValidator{Address: pk.Address(), PubKey: pk, Power: types.DefaultPower})
	agentInfos := make([]types.AgentInfo, 0)

	for _, val := range vals {
		agentInfos = append(agentInfos, types.AgentInfo{
			Address:  val.Address.String(),
			AgentUrl: types.DefaultAgentUrl,
			Name:     val.Address.String(),
		})
	}

	appState := types.GenesisAppState{
		Agents:   agentInfos,
		Manifest: types.DefaultStatement,
	}

	appStateJson, _ := json.MarshalIndent(appState, "", " ")
	genFile := appConfig.GenesisFile()
	appGenesis := &types.GenesisDoc{
		GenesisTime:     genesisTime,
		ChainID:         chainID,
		ConsensusParams: cmttypes.DefaultConsensusParams(),
		InitialHeight:   1,
		Validators:      vals,
		AppState:        appStateJson,
	}
	if err = types.ExportGenesisFile(appGenesis, genFile); err != nil {
		return fmt.Errorf("Failed to export genesis file %v", err)
	}
	app_config.WriteConfigFile(filepath.Join(appConfig.RootDir, "config", "config.toml"), appConfig)
	toPrint := newPrintInfo("", chainID, "", "", appGenesis.AppState)
	return displayInfo(toPrint)
}
