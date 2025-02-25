package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type NetworkConfig struct {
	HTTPUrl      string `mapstructure:"http_url"`
	HTTPPort     int    `mapstructure:"http_port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
	EnableTLS    bool   `mapstructure:"enable_tls"`
	AuthToken    string `mapstructure:"auth_token"`
	AgentUrl     string `mapstructure:"agent_url"`
}

func UpdateConfig(cfg *NetworkConfig) error {
	viper.Set("network.http_url", cfg.HTTPUrl)
	viper.Set("network.http_port", cfg.HTTPPort)
	viper.Set("network.read_timeout", cfg.ReadTimeout)
	viper.Set("network.write_timeout", cfg.WriteTimeout)
	viper.Set("network.idle_timeout", cfg.IdleTimeout)
	viper.Set("network.enable_tls", cfg.EnableTLS)
	viper.Set("network.auth_token", cfg.AuthToken)
	viper.Set("network.agent_url", cfg.AgentUrl)

	return viper.WriteConfig()
}

func LoadConfig(configPath string) (*NetworkConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg NetworkConfig
	if err := viper.UnmarshalKey("network", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
