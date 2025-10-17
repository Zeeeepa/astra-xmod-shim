package main

import (
	"astron-xmod-shim/internal/bootstrap"
	_ "astron-xmod-shim/internal/core/shimlet/shimlets" // Explicitly import plugin dependencies
	"log"
	"os"

	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "astron-xmod-shim",
		Short: "model serve shim",
		RunE:  runMw,
	}

	// Register config file parameter
	rootCmd.Flags().StringVarP(
		&configPath,
		"config", "c",
		"conf/base/conf.yaml",
		"Config file path",
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Startup failed: %v", err)
	}
}

// runMw Start middleware and register shutdown hooks
func runMw(cmd *cobra.Command, args []string) error {
	// 1. Validate config file
	if err := validateConfigFile(configPath); err != nil {
		return err
	}
	log.Printf("use cfg from: %s", configPath)

	// 2. Bootstrap
	if err := bootstrap.Init(configPath); err != nil {
		return err
	}

	// 3. Block and wait for shutdown signal
	waitForShutdownSignal()
	return nil
}

// validateConfigFile Verify config file exists
func validateConfigFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// waitForShutdownSignal Block and wait for shutdown signal
func waitForShutdownSignal() {
	// Wait for destructor waitGroup to complete callback
	bootstrap.WaitForShutDown()
	log.Println("receive shutdown signal waiting for resource release...")
}