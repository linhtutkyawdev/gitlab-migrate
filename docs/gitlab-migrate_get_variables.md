## gitlab-migrate get variables

Retrieve GitLab variables

### Synopsis

Retrieve CI/CD variables from GitLab groups or projects.
This command can fetch variables from:
- A specific group (using --group-id)
- A specific project (using --project-id)
- All projects within a group (using --group-id with --recursive)
The results can be saved to a file using the --output flag.

```
gitlab-migrate get variables [flags]
```

### Options

```
  -g, --group string     The GitLab group ID to retrieve projects for
  -h, --help             help for variables
  -p, --project string   The GitLab project ID to retrieve variables for
  -r, --recursive        Recursively retrieve variables from all projects in a group
```

### Options inherited from parent commands

```
  -c, --config string   Path to the config.yaml file (default: $HOME/config.yaml)
  -d, --destination     Uses the destination config instead of the source
  -o, --output string   Path to save the output as a JSON file
```

### SEE ALSO

* [gitlab-migrate get](gitlab-migrate_get.md)	 - Retrieve data from GitLab API using the provided config

###### Auto generated by spf13/cobra on 12-Dec-2024
