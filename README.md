# gitlab-migrate

🚀 **gitlab-migrate** is a powerful CLI tool designed to simplify the migration of GitLab projects using the [GitLab API](https://docs.gitlab.com/ee/api/). With is tool, it is easy to transfer projects, groups, and variables between GitLab instances or groups.

---

## Features

- 🛠 **Migrate GitLab projects**: Seamlessly transfer projects between GitLab instances or groups.
- 📂 **Manage project variables**: Get, set, and update environment variables.
- 🌟 **Made With Love By Lin Htut Kyaw**: Leverage GitLab’s robust API for efficient and secure migrations.
- 📚 **Built-in command documentation**: Explore individual commands for more flexibility.

---

## Installation

Download the prebuilt binary for your platform from the [Releases](https://gitlab.com/linhtutkyawdev/gitlab-migrate/-/releases) page.

1. Go to the [Releases](https://gitlab.com/linhtutkyawdev/gitlab-migrate/-/releases) page.
2. Download the appropriate binary for your operating system.
3. Rename the downloaded binary to `gitlab-migrate` and place it in a directory included in your system's `PATH`.

### Verifying the Installation

After installation, verify the installation by running:

```bash
gitlab-migrate --help
```

---

## Usage

### Basic Command

Run the following command to start using the CLI:

```bash
./gitlab-migrate --config /path/to/config.yaml
```

If no `--config` flag is provided, the tool will look for a `config.yaml` file in your home directory.

### Help Command

For a list of available commands and options:

```bash
./gitlab-migrate --help
```

---

## Commands Overview

The `gitlab-migrate` CLI tool includes the following commands:

| Command                           | Description                                    | Documentation File                                                           |
| -------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------- |
| `gitlab-migrate get groups`       | Retrieves and displays groups from GitLab      | [docs/gitlab-migrate_get_groups.md](docs/gitlab-migrate_get_groups.md)       |
| `gitlab-migrate get projects`     | Retrieves and displays projects from GitLab    | [docs/gitlab-migrate_get_projects.md](docs/gitlab-migrate_get_projects.md)   |
| `gitlab-migrate get variables`    | Retrieves project variables from GitLab        | [docs/gitlab-migrate_get_variables.md](docs/gitlab-migrate_get_variables.md) |
| `gitlab-migrate set variables`    | Sets or updates variables for a project        | [docs/gitlab-migrate_set_variables.md](docs/gitlab-migrate_set_variables.md) |
| `gitlab-migrate migrate variables`| Migrates variables between GitLab instances    | [docs/gitlab-migrate_migrate_variables.md](docs/gitlab-migrate_migrate_variables.md) |
| `gitlab-migrate mirror`            | Mirrors projects between GitLab instances      | [docs/gitlab-migrate_mirror.md](docs/gitlab-migrate_mirror.md)              |

### Common Command Examples

#### Get Commands
```bash
# Get all groups from source instance
gitlab-migrate get groups

# Get all groups from destination instance
gitlab-migrate get groups -d

# Get projects from a specific group
gitlab-migrate get projects -g GROUP_ID

# Get variables from a project
gitlab-migrate get variables -p PROJECT_ID

# Get variables recursively from all projects in a group
gitlab-migrate get variables -g GROUP_ID -r
```

#### Set Commands
```bash
# Set variables for a destination project
gitlab-migrate set variables -i input.json -P DEST_PROJECT_ID

# Set variables recursively for all projects in a destination group
gitlab-migrate set variables -i input.json -G DEST_GROUP_ID -r
```

#### Migrate Commands
```bash
# Migrate variables from source project to destination project
gitlab-migrate migrate variables -p SOURCE_PROJECT_ID -P DEST_PROJECT_ID

# Migrate variables from source group to destination group
gitlab-migrate migrate variables -g SOURCE_GROUP_ID -G DEST_GROUP_ID

# Migrate variables recursively from all projects in source group to destination group
gitlab-migrate migrate variables -g SOURCE_GROUP_ID -G DEST_GROUP_ID -r
```

#### Mirror Commands
```bash
# Mirror a single project
gitlab-migrate mirror -p <sourceProjectID> -P <targetProjectID>

# Mirror all projects in a group recursively
gitlab-migrate mirror -g <sourceGroupID> -G <targetGroupID>
```

### Flag Conventions
- Source identifiers use lowercase flags:
  - `-g` for source group ID
  - `-p` for source project ID
- Destination identifiers use uppercase flags:
  - `-G` for destination group ID
  - `-P` for destination project ID
- Other common flags:
  - `-d` use destination instance (for get commands)
  - `-r` recursive operation
  - `-i` input file path (for set commands)

---

## Configuration
The tool requires a configuration file (YAML format) with the following fields:
- `source_base_url`: The base URL of the source GitLab instance.
- `source_access_token`: The access token for the source GitLab API.
- `destination_base_url`: The base URL of the target GitLab instance.
- `destination_access_token`: The access token for the target GitLab API.

---

## Documentation

Explore detailed documentation for all commands in the [docs/](docs/) directory.

---

## Contributing

Contributions are welcome! Feel free to submit issues, fork the repository, and make a pull request.

1. Fork the project.
2. Create your feature branch:
   ```bash
   git checkout -b feature/new-feature
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add a new feature"
   ```
4. Push to the branch:
   ```bash
   git push origin feature/new-feature
   ```
5. Open a pull request.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Acknowledgements

This tool is developed by **Lin Htut Kyaw** and uses the GitLab API to provide seamless project migrations. Special thanks to all contributors!

---

## Feedback

We'd love to hear your thoughts! Feel free to open an issue or reach out directly.

Happy migrating! 🚀
