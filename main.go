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
	usage := `Batrak 2.1

	Usage:
		batrak -L [-n <issue>] [-c <count>]
		batrak -L -K [-c <count>]
		batrak -P
		batrak -M -n <issue> [<transition>]
		batrak -S -n <issue>
		batrak -T
		batrak -A -n <issue>
		batrak -C -n <issue>
		batrak -C -L -n <issue>
		batrak -C -L -n <issue> [-R] [<comment>]

	Commands:
		-L   List mode.
		-M   Move mode.
		-S   Start progress mode.
		-A   Assign issue mode.
		-C   Comments mode.
		-P   Projects mode.

	Options:
		-c <count>     Count of displayed issues [default: 10].
		-n <issue>     JIRA issue identifier.
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
	gojira.BaseUrl = config.JiraApiUrl

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
		projectsMode  = args["-P"].(bool)
	)

	switch {
	case startMode:
		err = handleStartMode(issueKey, hooks)

	case terminateMode:
		err = handleTerminateMode(hooks)

	case assignMode:
		err = handleAssignMode(issue, config.Username)

	case projectsMode:
		err = handleProjectsMode()

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

func handleProjectsMode() error {
	projects, err := gojira.GetProjects()
	if err != nil {
		return err
	}

	return displayProjects(projects)
}
