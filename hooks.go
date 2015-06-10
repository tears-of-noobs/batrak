package main

import (
	"fmt"
	"os/exec"
)

type Hooks struct {
	executables     map[string][]string
	userCredentials string
}

func NewHooks(config *Configuration) Hooks {
	return Hooks{
		executables:     config.Hooks,
		userCredentials: config.getUserCredentials(),
	}
}

func (hooks Hooks) Handle(action string, issueKey string) error {
	if executables, ok := hooks.executables[action]; ok {
		for _, executable := range executables {
			err := exec.Command(executable, issueKey, hooks.userCredentials).Run()
			if err != nil {
				return fmt.Errorf("hook '%s' failed: %s", executable, err)
			}
		}
	}

	return nil
}
