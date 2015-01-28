package main

import (
	"fmt"
	"gojira"
	"os"
	"os/user"

	"github.com/docopt/docopt.go"
)

var projectName string
var arguments map[string]interface{}

func init() {
	usage := `Batrak. 
	
	Usage:
		batrak (-L | --list) [-n NAME]
		batrak (-T | --transitions) [-n NAME]

	Commands:
		-L --list     List of last 10 issues assignee to logged username
		-T --transitions  List of available transitions for issue`

	arguments, _ = docopt.Parse(usage, nil, true, "Batrak 1.0", false)
}

func main() {
	//fmt.Printf("%s\n", arguments)
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configPath := fmt.Sprintf("%s/.batrakrc", usr.HomeDir)
	config, err := ReadConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gojira.Username = config.Username
	gojira.Password = config.Password
	gojira.BaseUrl = config.JiraApiUrl
	projectName = config.ProjectName

	user, err := gojira.Myself()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	//fmt.Printf("%v\n", user.Name)

	if arguments["-L"].(bool) == true || arguments["--list"].(bool) == true {
		if arguments["-n"].(bool) == true {
			jiraTag := arguments["NAME"].(string)
			PrintIssueByKey(jiraTag)
		} else {
			PrintIssues(user.Name)
		}
	}

	if arguments["-T"].(bool) == true || arguments["--transitions"].(bool) == true {
		if arguments["-n"].(bool) == true {
			jiraTag := arguments["NAME"].(string)
			PrintTransitionsOfIssue(jiraTag)
		}
	}

}
