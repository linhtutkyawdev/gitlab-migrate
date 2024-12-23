package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Version is the current version of gitlab-migrate
var Version = "v1.0.2"

var configPath string
var isDestination bool
var groupID string
var projectID string
var recursive bool
var outputFile string

// rootCmd represents the base command
// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gitlab-migrate",
	Short: "A CLI app to migrate GitLab projects using Gitlab API",
	Long: `gitlab-migrate is a command-line tool designed to migrate GitLab projects 
using the GitLab API and a configuration file written in YAML. It streamlines the 
process of transferring projects between GitLab instances or groups.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("==========================================")
		fmt.Println("🚀 Welcome to gitlab-migrate! 🚀")
		fmt.Println("==========================================")
		fmt.Println()
		fmt.Println("🔧 What does this tool do?")
		fmt.Println("  - Migrates GitLab projects between instances or groups.")
		fmt.Println("  - Powered by the GitLab API for seamless project transfers.")
		fmt.Println()
		fmt.Println("📄 How does it work?")
		fmt.Println("  1. Provide a YAML configuration file with your migration settings.")
		fmt.Println("  2. Specify the file path using the '--config' flag.")
		fmt.Println("  3. Sit back and let the magic happen!")
		fmt.Println()
		fmt.Println("📂 Default Configuration:")
		fmt.Println("  - If no '--config' flag is provided, it will look for 'config.yaml' in your home directory.")
		fmt.Println()
		fmt.Println("💡 Need help?")
		fmt.Println("  - Use '--help' for detailed usage and options.")
		fmt.Println()
		fmt.Println("==========================================")
		fmt.Println("🎉 Made with love by Lin Htut Kyaw")
		fmt.Println("==========================================")
	},
}

func Execute() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("gitlab-migrate {{.Version}}")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the config.yaml file (default: $HOME/config.yaml)")
	err := doc.GenMarkdownTree(rootCmd, "./docs")
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(NewMirrorCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
