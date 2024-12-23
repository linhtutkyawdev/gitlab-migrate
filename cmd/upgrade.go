package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

type Release struct {
	TagName string `json:"tag_name"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade gitlab-migrate to the latest version",
	Long: `Upgrade gitlab-migrate to the latest version from the official repository.
This command will check for the latest version and upgrade if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		currentVersion := Version // Version should be defined in root.go
		fmt.Printf("Current version: %s\n", currentVersion)

		// Get latest version from GitLab API
		resp, err := http.Get("https://gitlab.com/api/v4/projects/65329846/releases")
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}

		var releases []Release
		if err := json.Unmarshal(body, &releases); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		latestVersion := releases[0].TagName
		fmt.Printf("Latest version: %s\n", latestVersion)

		if currentVersion == latestVersion {
			fmt.Println("You are already using the latest version!")
			return
		}

		fmt.Printf("Upgrading to version %s...\n", latestVersion)

		// Determine the installation command based on the OS
		var updCmd *exec.Cmd
		switch runtime.GOOS {
		case "linux", "darwin":
			updCmd = exec.Command("go", "install", "gitlab.com/linhtutkyawdev/gitlab-migrate@"+latestVersion)
		case "windows":
			updCmd = exec.Command("go", "install", "gitlab.com/linhtutkyawdev/gitlab-migrate@"+latestVersion)
		default:
			fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
			return
		}

		output, err := updCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error upgrading: %v\n%s\n", err, string(output))
			return
		}

		fmt.Printf("Successfully upgraded to version %s!\n", latestVersion)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
