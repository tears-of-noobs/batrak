package main

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Username    string              `toml:"username"`
	Password    string              `toml:"password"`
	JiraApiUrl  string              `toml:"jira_api_url"`
	ProjectName string              `toml:"project_name"`
	Workflow    Workflow            `toml:"workflow"`
	Hooks       map[string][]string `toml:"hooks"`
	Filter      int                 `toml:"filter_id"`
}

type Workflow struct {
	AgileFields []string `toml:"agile_fields"`
	Stages      []Stage  `toml:"stage"`
}

type Stage struct {
	Name        string `toml:"name"`
	Order       int    `toml:"order"`
	KanbanOrder int    `toml:"kanban_order"`
}

func getConfig(filePath string) (*Configuration, error) {
	var config Configuration

	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return nil, err
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Configuration) Validate() error {
	switch {
	case config.Username == "":
		return errors.New("Username is empty")
	case config.Password == "":
		return errors.New("Password is empty")
	case config.JiraApiUrl == "":
		return errors.New("URL to Jira API is empty")
	case config.ProjectName == "":
		return errors.New(" Project name is empty")
	}

	return nil
}

func (config *Configuration) getUserCredentials() string {
	return fmt.Sprintf(
		"%s\x8E%s\x8E%s",
		config.Username, config.Password, config.JiraApiUrl,
	)
}

func loadWorkflow(path string, workflow *Workflow) error {
	_, err := toml.DecodeFile(path, &workflow)
	if err != nil {
		return err
	}

	return nil
}
