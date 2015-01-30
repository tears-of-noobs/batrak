package main

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Username    string   `toml:"username"`
	Password    string   `toml:"password"`
	JiraApiUrl  string   `toml:"jira_api_url"`
	ProjectName string   `toml:"project_name"`
	Workflow    Workflow `toml: "workflow"`
	Hooks       Hooks    `toml:"hooks"`
	Filter      int      `toml:"filter_id"`
}

type Hooks struct {
	PreStart  []string `toml:"pre_start"`
	PostStart []string `toml:"post_start"`
	PreStop   []string `toml:"pre_stop"`
	PostStop  []string `toml:"post_stop"`
}

type Workflow struct {
	AgileFields []string `toml:"agile_fields"`
	Stage       []Stage  `toml:"stage"`
}
type Stage struct {
	Name  string `toml: "name"`
	Order int    `toml: "order"`
}

func ReadConfig(filePath string) (*Configuration, error) {
	var config Configuration
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return nil, err
	}
	err = config.testConfig()
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Configuration) testConfig() error {
	if c.Username == "" {
		return errors.New("Username is empty")
	}
	if c.Password == "" {
		return errors.New("Password is empty")
	}
	if c.JiraApiUrl == "" {
		return errors.New("URL to Jira API is empty")
	}
	if c.ProjectName == "" {
		return errors.New(" Project name is empty")
	}
	return nil
}

func (c *Configuration) ExportToHook() string {
	return fmt.Sprintf("%s*%s*%s",
		c.Username,
		c.Password,
		c.JiraApiUrl)
}
