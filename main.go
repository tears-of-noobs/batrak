package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/tears-of-noobs/gojira"

	"github.com/docopt/docopt.go"
)

var config *Configuration
var tmpDir = "/tmp/batrak/"
var arguments map[string]interface{}

func init() {
	usage := `Batrak. 
	
	Usage:
		batrak (-L | --list) [-n NAME] [--count=<cnt>]
		batrak (-L | --list) [-C] [-n NAME]
		batrak (-M | --move) [-n NAME]
		batrak (-M | --move) [-n NAME] <TRANSITION>
		batrak (-S | --start) [-n NAME]
		batrak (-T | --terminate) [-n NAME]
		batrak (-A | --assign) [-n NAME]
		batrak (-C ) [-n NAME]
		batrak (-C ) [-R] [-n NAME] <COMMENTID>

	Commands:
		-L --list     List of last 10 issues assignee to logged username
		-M --move  List of available transitions for issue
		-S --start  Start progress on issue
	Options:
		--count=<cnt>  Count of prined issues [default: 10].`

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
	config, err = ReadConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err = os.Mkdir(tmpDir, 0777)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}
	gojira.Username = config.Username
	gojira.Password = config.Password
	gojira.BaseUrl = config.JiraApiUrl

	user, err := gojira.Myself()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	var jiraTag string
	if arguments["-n"].(bool) == true {
		jiraTag = arguments["NAME"].(string)
		tokens := strings.Split(jiraTag, "-")
		if len(tokens) < 2 {
			if !strings.Contains(jiraTag, config.ProjectName) {
				jiraTag = fmt.Sprintf("%s-%s", config.ProjectName, jiraTag)
			}

		}
	}

	if arguments["-L"].(bool) == true || arguments["--list"].(bool) == true {
		if arguments["-n"].(bool) == true {
			if arguments["-C"].(bool) == true {
				printComments(jiraTag)
			} else {
				printIssueByKey(jiraTag)
			}
		} else {
			cnt := arguments["--count"].(string)
			printIssues(user.Name, cnt)
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
				printTransitionsOfIssue(jiraTag)
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
	if arguments["-A"].(bool) == true || arguments["--assign"].(bool) == true {
		if arguments["-n"].(bool) == true {
			err := assignIssue(jiraTag)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			} else {
				fmt.Printf("Issue %s assignee to %s\n", jiraTag, config.Username)
			}
		}

	}
	if arguments["-C"].(bool) == true && arguments["-L"].(bool) == false {
		if arguments["-n"].(bool) == true {
			if arguments["-R"].(bool) == false {
				err := commentIssue(jiraTag)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err.Error())
				} else {
					fmt.Printf("Issue %s commented\n", jiraTag)
				}
			} else {
				commentId := arguments["<COMMENTID>"].(string)
				err := removeComment(jiraTag, commentId)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err.Error())
				} else {
					fmt.Printf("Comment %s removed\n", commentId)
				}

			}
		}

	}

}
