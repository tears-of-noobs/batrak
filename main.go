package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/tears-of-noobs/gojira"

	"github.com/docopt/docopt.go"
)

var projectName string
var workflow []Stage
var filterId int
var tmpDir = "/tmp/batrak/"
var arguments map[string]interface{}

func init() {
	usage := `Batrak. 
	
	Usage:
		batrak (-L | --list) [-n NAME]
		batrak (-L | --list) [-C] [-n NAME]
		batrak (-M | --move) [-n NAME]
		batrak (-M | --move) [-n NAME] <TRANSITION>
		batrak (-S | --start) [-n NAME]
		batrak (-T | --terminate) [-n NAME]

	Commands:
		-L --list     List of last 10 issues assignee to logged username
		-M --move  List of available transitions for issue
		-S --start  Start progress on issue`

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
	workflow = config.Workflow
	filterId = config.Filter

	user, err := gojira.Myself()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	var jiraTag string
	if arguments["-n"].(bool) == true {
		jiraTag = arguments["NAME"].(string)
		tokens := strings.Split(jiraTag, "-")
		if len(tokens) < 2 {
			if !strings.Contains(jiraTag, projectName) {
				jiraTag = fmt.Sprintf("%s-%s", projectName, jiraTag)
			}

		}
	}

	if arguments["-L"].(bool) == true || arguments["--list"].(bool) == true {
		if arguments["-n"].(bool) == true {
			if arguments["-C"].(bool) == true {
				printComments(jiraTag)
			} else {
				PrintIssueByKey(jiraTag)
			}
		} else {
			PrintIssues(user.Name)
		}
	}

	if arguments["-M"].(bool) == true || arguments["--move"].(bool) == true {
		if arguments["-n"].(bool) == true {
			if arguments["<TRANSITION>"] != nil {
				transId := arguments["<TRANSITION>"].(string)
				err := moveIssue(jiraTag, transId)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Issue moved")
				}

			} else {
				PrintTransitionsOfIssue(jiraTag)
			}
		}
	}
	if arguments["-T"].(bool) == true || arguments["--terminate"].(bool) == true {
		if arguments["-n"].(bool) == true {
			err := termProgress(jiraTag)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	if arguments["-S"].(bool) == true || arguments["--start"].(bool) == true {
		if arguments["-n"].(bool) == true {
			err := startProgress(jiraTag)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Issue started")
			}
		}

	}

}
