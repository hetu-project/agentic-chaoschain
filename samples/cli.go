package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nagara-stack/samples/config"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "hac-agent",
		Short: "HAC Node Command Line Tool",
		Long:  "HAC Agent is a command line tool for interacting with HAC blockchain nodes",
	}

	// Service start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start Agent HTTP service",
		Long:  "Start Agent HTTP service to handle blockchain interaction requests",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting Agent service...")
			startAgentService()
		},
	}

	// Register command
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register new Agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			url, _ := cmd.Flags().GetString("url")
			intro, _ := cmd.Flags().GetString("intro")

			// Load configuration
			cfg, err := config.LoadConfig("../config/")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Build base URL from configuration
			baseURL := fmt.Sprintf("%s:%d/api", cfg.HTTPUrl, cfg.HTTPPort)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client := NewHACNodeClient(cfg.AuthToken)
			client.baseURL = baseURL // Use URL from configuration

			err = client.Register(ctx, name, url, intro)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully registered Agent '%s'\n", name)
			return nil
		},
	}
	registerCmd.Flags().StringP("token", "t", "", "Authentication token (required)")
	registerCmd.Flags().StringP("name", "n", "", "Agent name (required)")
	registerCmd.Flags().StringP("url", "u", "", "Agent service URL (required)")
	registerCmd.Flags().StringP("intro", "i", "", "Agent introduction (required)")
	registerCmd.MarkFlagRequired("name")
	registerCmd.MarkFlagRequired("url")
	registerCmd.MarkFlagRequired("intro")

	// Proposal command
	proposalCmd := &cobra.Command{
		Use:   "propose",
		Short: "Submit new proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, _ := cmd.Flags().GetString("data")
			title, _ := cmd.Flags().GetString("title")

			// Load configuration
			cfg, err := config.LoadConfig("../config/")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Build base URL from configuration
			baseURL := fmt.Sprintf("%s:%d/api", cfg.HTTPUrl, cfg.HTTPPort)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client := NewHACNodeClient(cfg.AuthToken)
			client.baseURL = baseURL // Use URL from configuration

			err = client.PostProposal(ctx, data, title)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully submitted proposal: '%s'\n", title)
			return nil
		},
	}
	proposalCmd.Flags().StringP("token", "t", "", "Authentication token (required)")
	proposalCmd.Flags().StringP("data", "d", "", "Proposal content data (required)")
	proposalCmd.Flags().StringP("title", "T", "", "Proposal title (required)")
	proposalCmd.MarkFlagRequired("data")
	proposalCmd.MarkFlagRequired("title")

	// Add all commands to the root command
	rootCmd.AddCommand(startCmd, registerCmd, proposalCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
