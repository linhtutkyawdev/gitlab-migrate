package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gitlab.com/linhtutkyawdev/gitlab-migrate/utils"

	"github.com/spf13/cobra"
)

var inputFilePath string
var isSource bool
var destinationGroupID string
var destinationProjectID string

// setCmd is the parent command for "set" operations
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Update data in GitLab using the provided input",
	Long: `Set command allows you to update various data in GitLab using the API.
It can update variables and other settings based on your input file.
Use subcommands to specify what type of data you want to update.`,
}

// setVariablesCmd updates variables for projects
var setVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Update GitLab variables for projects based on the input file",
	Long: `Update CI/CD variables in GitLab projects or groups using an input file.
This command supports:
- Updating variables in a specific project
- Updating variables in a specific group
- Recursive updates to all projects within a group

The input file should contain the variables in JSON format.
Use --source flag for source GitLab instance or --destination for target instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig() // Pass the config file path here
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if inputFilePath == "" {
			fmt.Println("Error: Input file path is required.")
			return
		}

		if (destinationProjectID != "" && destinationGroupID != "") || (destinationProjectID == "" && destinationGroupID == "") {
			fmt.Println("Error: Either --destination-project or --destination-group must be provided.")
			return
		}

		if destinationGroupID != "" {
			if recursive {
				inputData, err := readRecursiveIputFile(inputFilePath)
				if err != nil {
					fmt.Printf("Error reading input file: %v\n", err)
					return
				}
				projects, err := fetchAllProjects(config)
				if err != nil {
					fmt.Printf("Error fetching projects: %v\n", err)
					return
				}

				for _, projectData := range inputData {
					projectName, ok := projectData["project_name"].(string)
					if !ok {
						fmt.Printf("Error: Project name is not in the correct format.\n")
						continue
					}
					projectID := findProjectIDByExactName(projects, projectName)
					if projectID == 0 {
						fmt.Printf("Error: Project %s not found in the destination.\n", projectName)
						continue
					}
					variables, ok := projectData["variables"].([]interface{})

					if !ok {
						fmt.Printf("Error: Variables for project %s are not in the correct format.\n", projectName)
						continue
					}

					createVariablesForProject(config, strconv.FormatInt(projectID, 10), variables)
				}
			} else {
				variables, err := readInputFile(inputFilePath)
				if err != nil {
					fmt.Printf("Error reading input file: %v\n", err)
					return
				}
				createVariablesForGroup(config, destinationGroupID, variables)
			}

		} else {
			variables, err := readInputFile(inputFilePath)
			if err != nil {
				fmt.Printf("Error reading input file: %v\n", err)
				return
			}
			createVariablesForProject(config, destinationProjectID, variables)
		}
	},
}

func readRecursiveIputFile(filePath string) (map[string]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	var parsedData map[string]map[string]interface{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %v", err)
	}

	return parsedData, nil
}

// readInputFile reads the input file for project variables
func readInputFile(filePath string) ([]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	var parsedData []interface{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %v", err)
	}

	return parsedData, nil
}

// fetchAllProjects retrieves all projects
func fetchAllProjects(config *utils.Config) ([]map[string]interface{}, error) {
	var allProjects []map[string]interface{}
	baseUrl := config.DestinationBaseURL
	accessToken := config.DestinationAccessToken
	page := 1

	if isSource {
		baseUrl = config.SourceBaseURL
		accessToken = config.SourceAccessToken
	}

	for {
		url := fmt.Sprintf("%s/api/v4/groups/%s/projects?per_page=100&page=%d", baseUrl, destinationGroupID, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("PRIVATE-TOKEN", accessToken)

		httpConfig := utils.NewDefaultConfig()
		httpConfig.SkipTLSVerification = true
		client := utils.CreateHTTPClient(httpConfig)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching projects: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error fetching projects: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		var projects []map[string]interface{}
		if err := json.Unmarshal(body, &projects); err != nil {
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

// findProjectIDByExactName searches for a project by exact name in the list of projects
func findProjectIDByExactName(projects []map[string]interface{}, projectName string) int64 {
	for _, project := range projects {
		if projectName == project["name"].(string) {
			return int64(project["id"].(float64))
		}
	}
	return 0
}

// createVariablesForProject updates variables for a specific project
func createVariablesForProject(config *utils.Config, projectID string, variables []interface{}) {
	var url string
	baseUrl := config.DestinationBaseURL
	accessToken := config.DestinationAccessToken

	if isSource {
		baseUrl = config.SourceBaseURL
		accessToken = config.SourceAccessToken
	}

	url = fmt.Sprintf("%s/api/v4/projects/%s/variables", baseUrl, projectID)

	for _, variable := range variables {
		payload, err := json.Marshal(variable)
		if err != nil {
			fmt.Printf("Error marshaling variable payload for project %s: %v\n", projectID, err)
			continue
		}

		// Use POST method to create the variable
		err = makeGitLabAPIRequest("POST", url, accessToken, string(payload))
		if err != nil {
			fmt.Printf("Error creating variable for project %s: %v\n", projectID, err)
		} else {
			fmt.Printf("Successfully created variable for project %s\n", projectID)
		}
	}
}

// createVariablesForGroup updates variables for a specific group
func createVariablesForGroup(config *utils.Config, groupID string, variables []interface{}) {
	var url string
	baseUrl := config.DestinationBaseURL
	accessToken := config.DestinationAccessToken

	if isSource {
		baseUrl = config.SourceBaseURL
		accessToken = config.SourceAccessToken
	}

	url = fmt.Sprintf("%s/api/v4/groups/%s/variables", baseUrl, groupID)

	for _, variable := range variables {
		payload, err := json.Marshal(variable)
		if err != nil {
			fmt.Printf("Error marshaling variable payload for group %s: %v\n", groupID, err)
			continue
		}

		// Use POST method to create the variable
		err = makeGitLabAPIRequest("POST", url, accessToken, string(payload))
		if err != nil {
			fmt.Printf("Error creating variable for group %s: %v\n", groupID, err)
		} else {
			fmt.Printf("Successfully created variable for group %s\n", groupID)
		}
	}
}

// makeGitLabAPIRequest makes an HTTP request to the GitLab API
func makeGitLabAPIRequest(method, url, token string, payload string) error {
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	httpConfig := utils.NewDefaultConfig()
	httpConfig.SkipTLSVerification = true
	client := utils.CreateHTTPClient(httpConfig)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned error status: %s", resp.Status)
	}

	return nil
}

func init() {
	// input file for setting variables
	setVariablesCmd.Flags().StringVarP(&inputFilePath, "input", "i", "", "Path to the input JSON file")
	setVariablesCmd.MarkFlagRequired("input")
	setVariablesCmd.Flags().StringVarP(&destinationProjectID, "destination-project", "P", "", "The destination project ID to set variables for")
	setVariablesCmd.Flags().StringVarP(&destinationGroupID, "destination-group", "G", "", "The destination group ID to set variables for")
	setVariablesCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively set variables from all projects in a group")
	setVariablesCmd.Flags().BoolVarP(&isSource, "source", "s", false, "Set variables to the source instance instead of the destination instance")

	setCmd.AddCommand(setVariablesCmd)
	rootCmd.AddCommand(setCmd)
}
