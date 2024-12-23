package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/linhtutkyawdev/gitlab-migrate/utils"
)

type MirrorCommand struct {
	sourceProjectID string
	targetProjectID string
	sourceGroupID   string
	targetGroupID   string
}

type MirrorPayload struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

func NewMirrorCommand() *cobra.Command {
	mc := &MirrorCommand{}
	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "Mirror GitLab projects between instances",
		Long: `Mirror GitLab projects between different instances.
Examples:
  - Mirror single project: mirror -p sourceProjectID -P targetProjectID
  - Mirror group projects: mirror -g sourceGroupID -G targetGroupID`,
		RunE: mc.Run,
	}

	// Add flags
	cmd.Flags().StringVarP(&mc.sourceProjectID, "source-project", "p", "", "Source project ID")
	cmd.Flags().StringVarP(&mc.targetProjectID, "target-project", "P", "", "Target project ID")
	cmd.Flags().StringVarP(&mc.sourceGroupID, "source-group", "g", "", "Source group ID")
	cmd.Flags().StringVarP(&mc.targetGroupID, "target-group", "G", "", "Target group ID")

	return cmd
}

func (mc *MirrorCommand) Run(cmd *cobra.Command, args []string) error {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Validate flags
	if (mc.sourceProjectID == "" && mc.sourceGroupID == "") ||
		(mc.targetProjectID == "" && mc.targetGroupID == "") {
		return fmt.Errorf("must specify either project IDs (-p, -P) or group IDs (-g, -G)")
	}

	if mc.sourceProjectID != "" && mc.targetProjectID != "" {
		return mc.mirrorProject(config, mc.sourceProjectID, mc.targetProjectID)
	}

	if mc.sourceGroupID != "" && mc.targetGroupID != "" {
		return mc.mirrorGroup(config, mc.sourceGroupID, mc.targetGroupID)
	}

	return nil
}

func (mc *MirrorCommand) mirrorProject(config *utils.Config, sourceID, targetID string) error {
	// Get source project details
	sourceURL := fmt.Sprintf("%s/api/v4/projects/%s", config.SourceBaseURL, sourceID)
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", config.SourceAccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get project details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get project details, status: %d", resp.StatusCode)
	}

	var project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return fmt.Errorf("failed to decode project details: %v", err)
	}

	// Check if credentials are set, if not prompt for them
	if config.AuthUser == "" || config.AuthPassword == "" {
		var username, password string
		fmt.Print("Enter mirror username: ")
		fmt.Scan(&username)
		fmt.Print("Enter mirror password: ")
		fmt.Scan(&password)
		config.AuthUser = username
		config.AuthPassword = password
		// Save updated config
		if err := writeConfigToFile(config, configPath); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
	}

	// Create mirror using the correct repository URL
	targetURL := fmt.Sprintf("%s/api/v4/projects/%s/remote_mirrors", config.DestinationBaseURL, targetID)
	payload := MirrorPayload{
		Enabled: true,
		URL:     strings.Replace(config.DestinationBaseURL, "https://", fmt.Sprintf("https://%s:%s@", config.AuthUser, config.AuthPassword), 1) + fmt.Sprintf("/%s.git", project.PathWithNamespace),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err = http.NewRequest("POST", targetURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", config.DestinationAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create mirror, status: %d", resp.StatusCode)
	}

	fmt.Printf("Successfully created mirror for project %s to %s\n", sourceID, targetID)
	return nil
}

func (mc *MirrorCommand) mirrorGroup(config *utils.Config, sourceGroupID, targetGroupID string) error {
	// Fetch all projects from source group
	sourceProjects, err := mc.fetchGroupProjects(config, sourceGroupID, true)
	if err != nil {
		return fmt.Errorf("failed to fetch source projects: %v", err)
	}

	// Fetch all projects from target group
	targetProjects, err := mc.fetchGroupProjects(config, targetGroupID, false)
	if err != nil {
		return fmt.Errorf("failed to fetch target projects: %v", err)
	}

	// Create a map of target project paths to IDs for quick lookup
	targetProjectMap := make(map[string]string)
	for _, project := range targetProjects {
		if namespace, ok := project["namespace"].(map[string]interface{}); ok {
			if name, ok := project["name"].(string); ok {
				path := fmt.Sprintf("%s/%s", namespace["name"].(string), name)
				if id, ok := project["id"].(float64); ok {
					targetProjectMap[path] = fmt.Sprintf("%.0f", id)
				}
			}
		}
	}

	// Process each source project
	for _, sourceProject := range sourceProjects {
		namespace, ok := sourceProject["namespace"].(map[string]interface{})
		if !ok {
			fmt.Printf("Warning: Could not get namespace for source project\n")
			continue
		}

		name, ok := sourceProject["name"].(string)
		if !ok {
			fmt.Printf("Warning: Could not get name for source project\n")
			continue
		}

		sourcePath := fmt.Sprintf("%s/%s", namespace["name"].(string), name)

		// Find corresponding target project
		targetID, exists := targetProjectMap[sourcePath]
		if !exists {
			fmt.Printf("Warning: Target project %s not found\n", sourcePath)
			continue
		}

		// Create mirror
		err := mc.mirrorProject(config, fmt.Sprintf("%.0f", sourceProject["id"].(float64)), targetID)
		if err != nil {
			fmt.Printf("Error mirroring project %s: %v\n", sourcePath, err)
			continue
		}
	}

	return nil
}

func (mc *MirrorCommand) fetchGroupProjects(config *utils.Config, groupID string, isSource bool) ([]map[string]interface{}, error) {
	var allProjects []map[string]interface{}
	baseURL := config.DestinationBaseURL
	accessToken := config.DestinationAccessToken
	if isSource {
		baseURL = config.SourceBaseURL
		accessToken = config.SourceAccessToken
	}

	page := 1
	for {
		url := fmt.Sprintf("%s/api/v4/groups/%s/projects?per_page=100&page=%d&include_subgroups=true", baseURL, groupID, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("PRIVATE-TOKEN", accessToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching projects: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error fetching projects: %s", resp.Status)
		}

		var projects []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
			return nil, fmt.Errorf("error parsing projects: %v", err)
		}

		if len(projects) == 0 {
			break
		}

		allProjects = append(allProjects, projects...)
		page++
	}

	return allProjects, nil
}
