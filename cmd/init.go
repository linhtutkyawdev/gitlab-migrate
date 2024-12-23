package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/linhtutkyawdev/gitlab-migrate/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd defines the "init" subcommand
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration by creating a config.yaml file",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		// If --config is not provided, default to the home directory
		if configPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Error finding home directory:", err)
				return
			}
			configPath = filepath.Join(homeDir, "config.yaml")
			fmt.Println("Defaulting to home directory:", configPath)

		} else if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			return
		}

		// Ask the user for configuration values
		fmt.Print("Enter Source Base URL: ")
		sourceBaseURL, _ := reader.ReadString('\n')
		sourceBaseURL = sanitizeInput(sourceBaseURL)

		fmt.Print("Enter Source Access Token: ")
		sourceAccessToken, _ := reader.ReadString('\n')
		sourceAccessToken = sanitizeInput(sourceAccessToken)

		fmt.Print("Enter Destination Base URL: ")
		destinationBaseURL, _ := reader.ReadString('\n')
		destinationBaseURL = sanitizeInput(destinationBaseURL)

		fmt.Print("Enter Destination Access Token: ")
		destinationAccessToken, _ := reader.ReadString('\n')
		destinationAccessToken = sanitizeInput(destinationAccessToken)

		// Create a Config struct
		config := &utils.Config{
			SourceBaseURL:          sourceBaseURL,
			SourceAccessToken:      sourceAccessToken,
			DestinationBaseURL:     destinationBaseURL,
			DestinationAccessToken: destinationAccessToken,
		}

		// Write the configuration to the specified file
		if err := writeConfigToFile(config, configPath); err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			return
		}

		fmt.Printf("Configuration saved successfully to %s\n", configPath)
	},
}

// Helper function to sanitize input
func sanitizeInput(input string) string {
	return input[:len(input)-1] // Remove newline character
}

// Helper function to write configuration to a file
func writeConfigToFile(config *utils.Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
