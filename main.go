package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/reconquest/executil-go"
	"github.com/tears-of-noobs/gojira"
)

func getArgs() (map[string]interface{}, error) {
	usage := `Batrak 3.3

   Batrak is a util for working with Jira using command line interface, is
summoned for increasing your efficiency in working with routine tasks.

Usage:
    batrak [options] -L [-K]
    batrak [options] -L <issue>
    batrak [options] -A <issue>
    batrak [options] -S <issue>
    batrak [options] -T <issue>
    batrak [options] -M <issue> [<transition>]
    batrak [options] -R <issue> <title>
    batrak [options] -C <issue>
    batrak [options] -C -L <issue>
    batrak [options] -C -L <issue> -R <comment>
    batrak [options] -D <issue>

Options:
    -L --list         List issues using specified filter. You can specify <issue>
                       identifier and see issue details.
                       Combine this flag with -K (--kanban) and
                       batrak will list issues in kanban board style.
      -c <count>      Limit amount of issues. [default: 10]
      -f <id>         Use specified filter identifier.
      -w --show-name  Show issue assignee username instead of "Display Name".
    -A --assign       Assign specified issue.
    -S --start        Start working on specified issue.
    -T --terminate    Stop working on specified issue.
    -M --move         Move specified issue or list available transitions.
    -D --delete       Delete specified issue.
    -R --rename       Change specified issue title to <title>. If new <title>
                       value starts with s/ then <title> will be used as
                       expression to sed with old title value as input.
    -C --comments     Create comment to specified issue.
                       Combine this flag with -L (--list) and
                       batrak will list comments to specified issue.
                       Combine this flag with -D (--delete) and
                       batrak will delete specified comment to specified issue.
  --config <path>     Use specified configuration file.
                       [default: $HOME/.batrakrc]
  -p <project>        Use specified project name instead of config.
  --workflow <path>   Rewrite configuration workflow using specified file.
  -v --version        Show version of the program.
`

	return docopt.Parse(os.ExpandEnv(usage), nil, true, "Batrak 3.3", false)
}

func main() {
	args, err := getArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	config, err := getConfig(args["--config"].(string))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if path, ok := args["--workflow"].(string); ok {
		err := loadWorkflow(path, &config.Workflow)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	gojira.Username = config.Username
	gojira.Password = config.Password
	gojira.BaseURL = config.JiraApiUrl

	if projectName, ok := args["-p"].(string); ok {
		config.ProjectName = projectName
	}

	hooks := NewHooks(config)

	var issueKey string
	var issue *gojira.Issue
	if args["<issue>"] != nil {
		issueKey = args["<issue>"].(string)

		issueKeyPieces := strings.Split(issueKey, "-")
		if len(issueKeyPieces) < 2 {
			if !strings.Contains(issueKey, config.ProjectName) {
				issueKey = fmt.Sprintf("%s-%s", config.ProjectName, issueKey)
			}
		} else if config.ProjectName == "" {
			config.ProjectName = issueKeyPieces[0]
		}

		issue, err = gojira.GetIssue(issueKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	if config.ProjectName == "" {
		fmt.Fprintln(
			os.Stderr,
			"project name is empty, "+
				"you can specify it in config, "+
				"pass -p flag or just specify an issue",
		)
		os.Exit(1)
	}

	var (
		listMode      = args["--list"].(bool)
		moveMode      = args["--move"].(bool)
		startMode     = args["--start"].(bool)
		terminateMode = args["--terminate"].(bool)
		assignMode    = args["--assign"].(bool)
		commentsMode  = args["--comments"].(bool)
		deleteMode    = args["--delete"].(bool)
		renameMode    = args["--rename"].(bool)
	)

	switch {
	case renameMode:
		var (
			title = args["<title>"].(string)
		)

		err = handleRenameMode(issue, title)

	case startMode:
		err = handleStartMode(issueKey, hooks)

	case terminateMode:
		err = handleTerminateMode(hooks)

	case deleteMode && !commentsMode:
		err = handleDeleteMode(issue)

	case assignMode:
		err = handleAssignMode(issue, config.Username)

	case commentsMode:
		commentID := ""
		if args["<comment>"] != nil {
			commentID = args["<comment>"].(string)
		}

		err = handleCommentsMode(issue, listMode, deleteMode, commentID)

	case listMode:
		if issue != nil {
			err = displayIssue(issue)
			break
		}

		var (
			kanbanMode     = args["-K"].(bool)
			rawLimit, _    = args["-c"].(string)
			rawFilterID, _ = args["-f"].(string)
			limit, _       = strconv.Atoi(rawLimit)
			filterID, _    = strconv.Atoi(rawFilterID)
			showName       = args["--show-name"].(bool)
		)

		err = handleListMode(
			filterID,
			limit,
			kanbanMode,
			config,
			showName,
		)

	case moveMode:
		transition := ""
		if args["<transition>"] != nil {
			transition = args["<transition>"].(string)
		}

		err = handleMoveMode(issue, transition)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func handleListMode(
	filterID int,
	limit int,
	kanbanMode bool,
	config *Configuration,
	showName bool,
) error {
	var (
		search *gojira.JiraSearchIssues
		err    error
	)

	if filterID == 0 {
		filterID = config.Filter
	}

	if filterID != 0 {
		search, err = searchIssuesByFilterID(filterID)
	} else {

		jiraUser, err := gojira.Myself()
		if err != nil {
			return err
		}

		search, err = searchIssues(
			jiraUser.Name, config.ProjectName, limit,
		)
	}

	if err != nil {
		return err
	}

	activeIssueKey, err := getActiveIssueKey()
	if err != nil {
		return err
	}

	if kanbanMode {
		workflowStages := config.Workflow.Stages
		sort.Sort(KanbanOrderSortableStages(workflowStages))

		board, err := NewKanbanBoard(search.Issues, workflowStages)
		if err != nil {
			return err
		}

		board.GenerateBoardData(activeIssueKey)

		board.Display()

		return nil
	} else {
		return displayIssues(
			sortIssuesByStatus(search.Issues, config.Workflow.Stages),
			activeIssueKey, showName,
			config.Workflow,
		)
	}
}

func handleMoveMode(
	issue *gojira.Issue,
	transition string,
) error {
	if transition == "" {
		transitions, err := issue.GetTransitions()
		if err != nil {
			return err
		}

		return displayTransitions(transitions)
	}

	transitionRequest := map[string]interface{}{
		"transition": map[string]string{
			"id": transition,
		},
	}

	jsonedRequest, err := json.Marshal(transitionRequest)
	if err != nil {
		return err
	}

	err = issue.SetTransition(bytes.NewBuffer(jsonedRequest))
	if err != nil {
		return err
	}

	fmt.Printf("Issue %s moved\n", issue.Key)

	return nil
}

func handleTerminateMode(
	hooks Hooks,
) error {
	activeIssueKey, err := getActiveIssueKey()
	if err != nil {
		return err
	}

	if activeIssueKey == "" {
		return fmt.Errorf("You have not started issue")
	}

	issue, err := gojira.GetIssue(activeIssueKey)
	if err != nil {
		return err
	}

	err = stopProgress(issue, hooks)

	return err
}

func handleStartMode(
	issueKey string,
	hooks Hooks,
) error {
	activeIssueKey, err := getActiveIssueKey()
	if err != nil {
		return err
	}

	if activeIssueKey != "" {
		return fmt.Errorf(
			"You already have started have issue (%s)",
			activeIssueKey,
		)
	}

	err = startProgress(issueKey, hooks)
	if err != nil {
		return err
	}

	fmt.Printf("Issue %s started\n", issueKey)

	return nil
}

func handleDeleteMode(issue *gojira.Issue) error {
	err := issue.Delete()
	if err != nil {
		return err
	}

	fmt.Printf("Issue %s deleted\n", issue.Key)

	return nil
}

func handleAssignMode(
	issue *gojira.Issue,
	username string,
) error {
	err := issue.Assignee(username)
	if err != nil {
		return err
	}

	fmt.Printf("Issue %s successfully assigned to '%s'\n", issue.Key, username)

	return nil
}

func handleCommentsMode(
	issue *gojira.Issue, listMode bool, deleteMode bool, rawCommentID string,
) error {
	switch {
	case deleteMode:
		commentID, err := strconv.ParseInt(rawCommentID, 10, 64)
		if err != nil {
			return err
		}

		err = issue.DeleteComment(commentID)
		if err != nil {
			return nil
		}

		fmt.Printf("Comment #%d of issue %s deleted\n", commentID, issue.Key)

		return nil

	case listMode:
		comments, err := issue.GetComments()
		if err != nil {
			return err
		}

		return displayComments(comments)

	default:
		err := addComment(issue)
		if err != nil {
			return err
		}

		fmt.Printf("Issue %s successfully commented\n", issue.Key)

		return nil
	}
}

func handleRenameMode(
	issue *gojira.Issue, title string,
) error {
	if strings.HasPrefix(title, "s/") {
		cmd := exec.Command("sed", "-r", title)
		cmd.Stdin = bytes.NewBufferString(issue.Fields.Summary)
		seded, _, err := executil.Run(cmd)
		if err != nil {
			return err
		}

		title = string(seded)
	}

	err := issue.SetSummary(title)
	if err != nil {
		return err
	}

	fmt.Println(issue.Key + " successfully renamed to: " + title)

	return nil
}
