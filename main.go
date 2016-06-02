package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/tears-of-noobs/gojira"
)

func getArgs() (map[string]interface{}, error) {
	usage := `Batrak 3.0

   Batrak is a util for working with Jira using command line interface, is
summoned for increasing your efficiency in working with routine tasks.

Usage:
	batrak -L [-K]
	batrak -L <issue>
	batrak -M <issue> [<transition>]
	batrak -S <issue>
	batrak -T <issue>
	batrak -A <issue>
	batrak -C <issue>
	batrak -C -L <issue>
	batrak -C -L <issue> -R <comment>
	batrak -R -n <issue>

Options:
    -L --list       List issues using specified filter. You can specify <issue>
                      identifier and see issue details.
				      Combine this flag with -K (--kanban) and
				        batrak will list issues in kanban board style.
      -c <count>      Limit amount of issues. [default: 10]
    -M --move       Move specified issue or list available transitions.
    -S --start      Start working on specified issue.
    -R --remove     Delete specified issue.
    -T --terminate  Stop working on specified issue.
    -A --assign     Assign specified issue.
    -C --comments   Create comment to specified issue.
				      Combine this flag with -L (--list) and
				         batrak will list comments to specified issue.
				      Combine this flag with -R (--remove) and
				         batrak will remove specified comment to specified issue.
`

	return docopt.Parse(usage, nil, true, "Batrak 2.1", false)
}

func main() {
	args, err := getArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	config, err := getConfig(filepath.Join(os.Getenv("HOME"), ".batrakrc"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	gojira.Username = config.Username
	gojira.Password = config.Password
	gojira.BaseURL = config.JiraApiUrl

	hooks := NewHooks(config)

	var issueKey string
	var issue *gojira.Issue
	if args["-n"] != nil {
		issueKey = args["-n"].(string)

		issueKeyPieces := strings.Split(issueKey, "-")
		if len(issueKeyPieces) < 2 {
			if !strings.Contains(issueKey, config.ProjectName) {
				issueKey = fmt.Sprintf("%s-%s", config.ProjectName, issueKey)
			}
		}

		issue, err = gojira.GetIssue(issueKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	var (
		listMode      = args["-L"].(bool)
		moveMode      = args["-M"].(bool)
		startMode     = args["-S"].(bool)
		terminateMode = args["-T"].(bool)
		assignMode    = args["-A"].(bool)
		commentsMode  = args["-C"].(bool)
		removeMode    = args["-R"].(bool)
	)

	switch {
	case startMode:
		err = handleStartMode(issueKey, hooks)

	case terminateMode:
		err = handleTerminateMode(hooks)

	case removeMode && !commentsMode:
		err = handleRemoveMode(issue)

	case assignMode:
		err = handleAssignMode(issue, config.Username)

	case commentsMode:
		commentID := ""
		if args["<comment>"] != nil {
			commentID = args["<comment>"].(string)
		}

		err = handleCommentsMode(issue, listMode, removeMode, commentID)

	case listMode:
		if issue != nil {
			err = displayIssue(issue)
			break
		}

		var (
			issueListLimit, _ = strconv.Atoi(args["-c"].(string))
			kanbanMode        = args["-K"].(bool)
		)

		err = handleListMode(
			issue,
			issueListLimit,
			kanbanMode,
			config,
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
	issue *gojira.Issue,
	issueListLimit int,
	kanbanMode bool,
	config *Configuration,
) error {
	var (
		search *gojira.JiraSearchIssues
		err    error
	)

	if config.Filter != 0 {
		search, err = searchIssuesByFilterID(config.Filter)
	} else {

		jiraUser, err := gojira.Myself()
		if err != nil {
			return err
		}

		search, err = searchIssues(
			jiraUser.Name, config.ProjectName, issueListLimit,
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
			activeIssueKey,
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

func handleRemoveMode(issue *gojira.Issue) error {
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
	issue *gojira.Issue, listMode bool, removeMode bool, rawCommentID string,
) error {
	switch {
	case removeMode:
		commentID, err := strconv.ParseInt(rawCommentID, 10, 64)
		if err != nil {
			return err
		}

		err = issue.DeleteComment(commentID)
		if err != nil {
			return nil
		}

		fmt.Printf("Comment #%d of issue %s removed\n", commentID, issue.Key)

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
