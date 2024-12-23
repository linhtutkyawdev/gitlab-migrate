package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/linhtutkyawdev/gitlab-migrate/utils"

	"github.com/spf13/cobra"
)

// Constants for API and pagination
const (
	defaultPerPage = 100
	maxRetries     = 3
	retryDelay     = 2 * time.Second
)

// getCmd is the parent command for "get" operations
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve data from GitLab API using the provided config",
	Long: `Get command allows you to retrieve various data from GitLab using the API.
It can fetch groups, projects, and variables based on your configuration.
Use subcommands to specify what type of data you want to retrieve.`,
}

// getProjectsCmd retrieves projects
var getProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Retrieve GitLab projects",
	Long: `Retrieve a list of GitLab projects based on your configuration.
This command will fetch all accessible projects from the specified GitLab instance.
The results can be saved to a file using the --output flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			log.Printf("Error loading config: %v", err)
			return
		}

		var projects interface{}
		if groupID != "" {
			projects = getProjectsForGroup(config, groupID)
		} else {
			projects = executeGitLabAPIRequest(config.SourceBaseURL, config.SourceAccessToken, "projects")
		}

		if err := utils.EnsureDataDir(); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		if outputFile == "" {
			outputFile = utils.GenerateOutputFileName("projects", groupID, "", isDestination, false)
		}

		if err := saveOutputToFile(projects, outputFile); err != nil {
			log.Printf("Error saving output to file: %v", err)
			return
		}
	},
}

func saveOutputToFile(data interface{}, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	log.Printf("Successfully saved output to %s", filePath)
	return nil
}

// getProjectsForGroup retrieves projects for a specific group
func getProjectsForGroup(config *utils.Config, groupID string) []map[string]interface{} {
	var url string
	var accessToken string
	if isDestination {
		url = fmt.Sprintf("%s/api/v4/groups/%s/projects", config.DestinationBaseURL, groupID)
		accessToken = config.DestinationAccessToken
	} else {
		url = fmt.Sprintf("%s/api/v4/groups/%s/projects", config.SourceBaseURL, groupID)
		accessToken = config.SourceAccessToken
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for group %s: %v", groupID, err)
		return nil
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)

	httpConfig := utils.NewDefaultConfig()
	httpConfig.SkipTLSVerification = true
	client := utils.CreateHTTPClient(httpConfig)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching projects for group %s: %v", groupID, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading projects response for group %s: %v", groupID, err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching projects for group %s: %s", groupID, body)
		return nil
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(body, &projects); err != nil {
		log.Printf("Error parsing projects JSON for group %s: %v", groupID, err)
		return nil
	}

	return projects
}

// getGroupsCmd retrieves groups
var getGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Retrieve GitLab groups",
	Long: `Retrieve a list of GitLab groups based on your configuration.
This command will fetch all accessible groups from the specified GitLab instance.
The results can be saved to a file using the --output flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			log.Printf("Error loading config: %v", err)
			return
		}

		var accessToken string
		var baseURL string
		if isDestination {
			accessToken = config.DestinationAccessToken
			baseURL = config.DestinationBaseURL
		} else {
			accessToken = config.SourceAccessToken
			baseURL = config.SourceBaseURL
		}

		groups := executeGitLabAPIRequest(baseURL, accessToken, "groups")

		if err := utils.EnsureDataDir(); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		if outputFile == "" {
			outputFile = utils.GenerateOutputFileName("groups", "", "", isDestination, false)
		}

		if err := saveOutputToFile(groups, outputFile); err != nil {
			log.Printf("Error saving output to file: %v", err)
			return
		}
	},
}

// getVariablesCmd retrieves variables based on group or project
var getVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Retrieve GitLab variables",
	Long: `Retrieve CI/CD variables from GitLab groups or projects.
This command can fetch variables from:
- A specific group (using --group-id)
- A specific project (using --project-id)
- All projects within a group (using --group-id with --recursive)
The results can be saved to a file using the --output flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			log.Printf("Error loading config: %v", err)
			return
		}

		if groupID == "" && projectID == "" {
			log.Println("Error: Either --group or --project must be provided.")
			return
		}

		if err := utils.EnsureDataDir(); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		if outputFile == "" {
			outputFile = utils.GenerateOutputFileName("variables", groupID, projectID, isDestination, recursive)
		}

		if groupID != "" {
			if recursive {
				variablesByProject := getAllVariablesForGroupProjects(config, groupID)
				if err := saveOutputToFile(variablesByProject, outputFile); err != nil {
					log.Printf("Error saving output to file: %v", err)
					return
				}
			} else {
				variables := getVariablesForGroup(config, groupID)
				if err := saveOutputToFile(variables, outputFile); err != nil {
					log.Printf("Error saving output to file: %v", err)
					return
				}
			}
		} else if projectID != "" {
			if recursive {
				log.Println("Error: Recursive mode is not supported for individual projects.")
				return
			}
			variables := getVariablesForProject(config, projectID)
			if err := saveOutputToFile(variables, outputFile); err != nil {
				log.Printf("Error saving output to file: %v", err)
				return
			}
		}
	},
}

// getAllVariablesForGroupProjects retrieves variables for all projects in a group
func getAllVariablesForGroupProjects(config *utils.Config, groupID string) map[string]map[string]interface{} {
	projects := getProjectsForGroup(config, groupID)

	var variablesByProject = make(map[string]map[string]interface{})
	for _, project := range projects {
		projectID := int(math.Round(project["id"].(float64)))
		projectName := project["name"].(string)

		// Fetch variables for the project
		variables := getVariablesForProject(config, fmt.Sprintf("%d", projectID))

		// Create an entry combining the project name and its variables
		variablesByProject[fmt.Sprintf("%d", projectID)] = map[string]interface{}{
			"project_name": projectName,
			"variables":    variables,
		}
	}
	return variablesByProject
}

// loadConfig loads the configuration from the specified or default location
func loadConfig() (*utils.Config, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to find home directory: %v", err)
		}
		configPath = filepath.Join(homeDir, "config.yaml")
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %v", configPath, err)
	}
	return config, nil
}

// executeGitLabAPIRequest makes a request to the GitLab API for a specific resource
func executeGitLabAPIRequest(baseURL, token, resource string) interface{} {
	client := &http.Client{Timeout: 30 * time.Second}

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("Retrying request (attempt %d/%d)...", retry+1, maxRetries)
			time.Sleep(retryDelay)
		}

		url := fmt.Sprintf("%s/api/v4/%s?per_page=%d", baseURL, resource, defaultPerPage)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}

		req.Header.Set("PRIVATE-TOKEN", token)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making request: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("API request failed with status %d: %s", resp.StatusCode, string(body))
			if retry == maxRetries-1 {
				return nil
			}
			continue
		}

		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding response: %v", err)
			continue
		}

		return result
	}

	log.Printf("Failed to execute GitLab API request after %d attempts", maxRetries)
	return nil
}

// getVariablesForGroup retrieves variables for a specific GitLab group
func getVariablesForGroup(config *utils.Config, groupID string) []map[string]interface{} {
	var url string
	var accessToken string
	if isDestination {
		url = fmt.Sprintf("%s/api/v4/groups/%s/variables", config.DestinationBaseURL, groupID)
		accessToken = config.DestinationAccessToken
	} else {
		url = fmt.Sprintf("%s/api/v4/groups/%s/variables", config.SourceBaseURL, groupID)
		accessToken = config.SourceAccessToken
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for group %s: %v", groupID, err)
		return nil
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)

	httpConfig := utils.NewDefaultConfig()
	httpConfig.SkipTLSVerification = true
	client := utils.CreateHTTPClient(httpConfig)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching variables for group %s: %v", groupID, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading variables response for group %s: %v", groupID, err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching variables for group %s: %s", groupID, body)
		return nil
	}

	var variables []map[string]interface{}
	if err := json.Unmarshal(body, &variables); err != nil {
		log.Printf("Error parsing variables JSON for group %s: %v", groupID, err)
		return nil
	}

	return variables
}

// getVariablesForProject retrieves variables for a specific GitLab project
func getVariablesForProject(config *utils.Config, projectID string) []map[string]interface{} {

	var url string
	var accessToken string
	if isDestination {
		url = fmt.Sprintf("%s/api/v4/projects/%s/variables", config.DestinationBaseURL, projectID)
		accessToken = config.DestinationAccessToken
	} else {
		url = fmt.Sprintf("%s/api/v4/projects/%s/variables", config.SourceBaseURL, projectID)
		accessToken = config.SourceAccessToken
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for project %s: %v", projectID, err)
		return nil
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)

	httpConfig := utils.NewDefaultConfig()
	httpConfig.SkipTLSVerification = true
	client := utils.CreateHTTPClient(httpConfig)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching variables for project %s: %v", projectID, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading variables response for project %s: %v", projectID, err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching variables for project %s: %s", projectID, body)
		return nil
	}

	var variables []map[string]interface{}
	if err := json.Unmarshal(body, &variables); err != nil {
		log.Printf("Error parsing variables JSON for project %s: %v", projectID, err)
		return nil
	}

	return variables
}

func init() {
	// print the output to a file
	getCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Path to save the output as a JSON file")
	// get from destination rather than source
	getCmd.PersistentFlags().BoolVarP(&isDestination, "destination", "d", false, "Uses the destination config instead of the source")
	// filter projects by group
	getProjectsCmd.Flags().StringVarP(&groupID, "group", "g", "", "The GitLab group ID to retrieve projects for")
	// filter variables by project
	getVariablesCmd.Flags().StringVarP(&projectID, "project", "p", "", "The GitLab project ID to retrieve variables for")
	// filter variables by group
	getVariablesCmd.Flags().StringVarP(&groupID, "group", "g", "", "The GitLab group ID to retrieve projects for")

	// recursively retrieve variables from all projects
	getVariablesCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively retrieve variables from all projects in a group")

	// Register subcommands
	getCmd.AddCommand(getGroupsCmd)
	getCmd.AddCommand(getProjectsCmd)
	getCmd.AddCommand(getVariablesCmd)

	// Add "get" to the root command
	rootCmd.AddCommand(getCmd)
}
