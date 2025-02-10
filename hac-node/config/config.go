package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
)

type HACAppConfig struct {
	Home           string `mapstructure:"-"`
	TimeoutCommit  uint64 `mapstructure:"-"`
	AgentUrl       string `mapstructure:"agent_url"`
	ServiceAddress string `mapstructure:"service_address"`
}

func DefaultHACAppConfig(home string) *HACAppConfig {
	return &HACAppConfig{
		Home:     home,
		AgentUrl: "http://127.0.0.1:3000",
	}

}
func NewHACAppConfig(home string) *HACAppConfig {
	return &HACAppConfig{
		Home:     home,
		AgentUrl: "http://127.0.0.1:3000",
	}
}

func GWeiPerPower(height uint64) uint64 {
	return 1000000000
}

func PowerPerStake(stake uint64, height uint64) int64 {
	return int64(stake / GWeiPerPower(height))
}

type Config struct {
	*config.Config `mapstructure:",squash"`

	App *HACAppConfig `mapstructure:"app"`
}

func DefaultConfig(home string) *Config {
	if len(home) == 0 {
		home = os.ExpandEnv("$HOME/.hac")
	}
	config := &Config{
		DefaultHACCometConfig(),
		NewHACAppConfig(home),
	}
	config.RootDir = home
	_ = os.MkdirAll(home+"/config", 0755)
	return config
}

func InitializeOwner(home string) (owner string) {
	priv, _ := eth_crypto.GenerateKey()
	d := eth_crypto.FromECDSA(priv)
	key := hex.EncodeToString(d)

	err := os.WriteFile(home+"/config/owner_priv_key", []byte(key), 0644)
	if err != nil {
		fmt.Println("Error writing private key to file:", err)
		return
	}
	owner = eth_crypto.PubkeyToAddress(priv.PublicKey).Hex()
	return
}

func NewHACConfig(home string) *Config {
	if len(home) == 0 {
		home = os.ExpandEnv("$HOME/.hac")
	}
	_ = os.MkdirAll(home+"/config", 0755)
	config := &Config{
		DefaultHACCometConfig(),
		NewHACAppConfig(home),
	}
	config.RootDir = home
	return config
}

func InitializeNodeValidatorFiles(config *Config, privKey crypto.PrivKey) (nodeID string, pk crypto.PubKey, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}
	nodeID = string(nodeKey.ID())

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvKeyFile), err)
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(pvStateFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvStateFile), err)
	}

	var filePV *privval.FilePV
	if privKey == nil {
		filePV = privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	} else {
		filePV = privval.NewFilePV(privKey, pvKeyFile, pvStateFile)
		filePV.Save()
	}
	pukey, err := filePV.GetPubKey()
	if err != nil {
		return "", nil, err
	}

	return nodeID, pukey, nil
}

func InitializeNodeOnly(config *Config) {
	_, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(pvStateFile), 0o777); err != nil {
		return
	}
	privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	os.Remove(pvKeyFile)
}

func DefaultHACCometConfig() *config.Config {
	cometConfig := config.DefaultConfig()
	cometConfig.Consensus.TimeoutPropose = time.Second * 10
	cometConfig.Consensus.TimeoutPrevote = time.Second * 1
	cometConfig.Consensus.TimeoutPrecommit = time.Second * 1
	cometConfig.Consensus.TimeoutCommit = time.Millisecond * 1200
	return cometConfig
}
