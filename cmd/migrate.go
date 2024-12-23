package cmd

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"gitlab.com/linhtutkyawdev/gitlab-migrate/utils"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate GitLab resources between instances",
	Long: `Migrate GitLab resources between different GitLab instances or groups.
This command helps you transfer various GitLab resources including:
- CI/CD Variables
- Project settings
- Group settings

Use the appropriate subcommand to specify what you want to migrate.`,
}

var migrateVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Migrate variables between GitLab instances",
	Long: `Migrate CI/CD variables between GitLab instances, groups, or projects.
This command supports:
- Migrating variables from one group to another
- Migrating variables from one project to another
- Recursive migration of variables for all projects in a group

Required flags:
- Source: Use either -g (group ID) or -p (project ID)
- Destination: Use either --destination-group or --destination-project`,
	Run: func(cmd *cobra.Command, args []string) {
		if (groupID == "" && projectID == "") || (destinationGroupID == "" && destinationProjectID == "") {
			log.Println("Error: Source and destination IDs must be provided using one of:")
			log.Println("  - Source group (-g) and destination group (--destination-group)")
			log.Println("  - Source project (-p) and destination project (--destination-project)")
			return
		}

		// Load configuration
		config, err := loadConfig()
		if err != nil {
			log.Printf("Error loading config: %v", err)
			return
		}

		if err := utils.EnsureDataDir(); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		// Get source variables
		var sourceVars interface{}
		if groupID != "" {
			if recursive {
				sourceVars = getAllVariablesForGroupProjects(config, groupID)
			} else {
				sourceVars = getVariablesForGroup(config, groupID)
			}
		} else {
			sourceVars = getVariablesForProject(config, projectID)
		}

		// Save source variables to file (for reference)
		sourceFile := utils.GenerateOutputFileName("variables", groupID, projectID, false, recursive)
		if err := saveOutputToFile(sourceVars, sourceFile); err != nil {
			log.Printf("Error saving source variables: %v", err)
			return
		}

		// Create variables in destination
		if groupID != "" {
			if recursive {
				log.Printf("Migrating variables recursively from group %s to group %s", groupID, destinationGroupID)
				sourceVarsMap, ok := sourceVars.(map[string]map[string]interface{})
				if !ok {
					log.Printf("Error: Invalid source variables format")
					return
				}

				// Get destination projects to map names to IDs
				destProjects, err := fetchAllProjects(config)
				if err != nil {
					log.Printf("Error fetching destination projects: %v", err)
					return
				}

				for sourceProjectID, projectData := range sourceVarsMap {
					projectName, ok := projectData["project_name"].(string)
					if !ok {
						log.Printf("Error: Project name not found for project %s", sourceProjectID)
						continue
					}

					// Find the corresponding project in destination
					destProjectID := findProjectIDByExactName(destProjects, projectName)
					if destProjectID == 0 {
						log.Printf("Warning: Project %s not found in destination group", projectName)
						continue
					}

					vars, ok := projectData["variables"].([]map[string]interface{})
					if !ok {
						log.Printf("Error: Invalid variables format for project %s", projectName)
						continue
					}

					// Convert []map[string]interface{} to []interface{}
					interfaceVars := make([]interface{}, len(vars))
					for i, v := range vars {
						interfaceVars[i] = v
					}

					log.Printf("Migrating variables for project %s (ID: %d)", projectName, destProjectID)
					createVariablesForProject(config, strconv.FormatInt(destProjectID, 10), interfaceVars)
				}
			} else {
				log.Printf("Migrating variables from group %s to group %s", groupID, destinationGroupID)
				vars, ok := sourceVars.([]map[string]interface{})
				if !ok {
					log.Printf("Error: Invalid source variables format")
					return
				}
				// Convert []map[string]interface{} to []interface{}
				interfaceVars := make([]interface{}, len(vars))
				for i, v := range vars {
					interfaceVars[i] = v
				}
				createVariablesForGroup(config, destinationGroupID, interfaceVars)
			}
		} else {
			log.Printf("Migrating variables from project %s to project %s", projectID, destinationProjectID)
			vars, ok := sourceVars.([]map[string]interface{})
			if !ok {
				log.Printf("Error: Invalid source variables format")
				return
			}
			// Convert []map[string]interface{} to []interface{}
			interfaceVars := make([]interface{}, len(vars))
			for i, v := range vars {
				interfaceVars[i] = v
			}
			createVariablesForProject(config, destinationProjectID, interfaceVars)
		}

		log.Println("Variables migration completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateVariablesCmd)

	// Add flags for source IDs
	migrateVariablesCmd.Flags().StringVarP(&groupID, "group", "g", "", "Source group ID")
	migrateVariablesCmd.Flags().StringVarP(&projectID, "project", "p", "", "Source project ID")
	migrateVariablesCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively migrate variables from all projects in a group")

	// Add flags for destination IDs
	migrateVariablesCmd.Flags().StringVarP(&destinationGroupID, "destination-group", "G", "", "Destination group ID")
	migrateVariablesCmd.Flags().StringVarP(&destinationProjectID, "destination-project", "P", "", "Destination project ID")
}
