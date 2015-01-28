package main

import (
	"errors"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	JiraApiUrl  string `toml:"jira_api_url"`
	ProjectName string `toml:"project_name"`
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
