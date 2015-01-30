package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tears-of-noobs/gojira"
)

func statusOrder(iss gojira.Issue) int {
	if len(config.Workflow.Stage) == 0 {
		return 1
	}
	for _, stage := range config.Workflow.Stage {
		if stage.Name == iss.Fields.Status.Name {
			return stage.Order
		}
	}

	fmt.Println("Unknow workflow stage:", iss.Fields.Status.Name)
	return -1
}

type sortByStatus []gojira.Issue

func (v sortByStatus) Len() int      { return len(v) }
func (v sortByStatus) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortByStatus) Less(i, j int) bool {
	return statusOrder(v[i]) < statusOrder(v[j])
}

func assignIssue(issueKey string) error {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		return err
	}
	err = issue.Assignee(config.Username)
	if err != nil {
		return err
	}
	return nil
}

func commentIssue(issueKey string) error {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Write your comment by one line")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	var b = []byte(fmt.Sprintf(`{ "body": "%s" }`, strings.Trim(text, "\n")))
	_, err = issue.SetComment(bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	return nil

}

func removeComment(issueKey, commentId string) error {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		fmt.Println(err)
	}
	cId, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return err
	}
	err = issue.DeleteComment(cId)
	if err != nil {
		return nil
	}

	return nil
}

func PrintIssues(user, cnt string) {
	var result []byte
	var err error
	if config.Filter == 0 {
		searchString := "project%20%3D%20" + config.ProjectName +
			"%20AND%20assignee%20%3D%20" + user + "%20order%20by%20updated%20DESC" +
			"&fields=key,summary,status,assignee&maxResults=" + cnt
		result, err = gojira.RawSearch(searchString)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		result, err = gojira.FilterSearch(config.Filter)
		if err != nil {
			fmt.Println(err)
		}
	}
	var jiraIssues gojira.JiraSearchIssues
	err = json.Unmarshal(result, &jiraIssues)
	if err != nil {
		fmt.Println(err)
	}

	sort.Sort(sortByStatus(jiraIssues.Issues))
	for _, issue := range jiraIssues.Issues {
		var started string
		if checkActive(issue.Key) {
			started = "*"
		}
		fmt.Printf("%2s %10s %15s %13s %s\n", started, issue.Key,
			issue.Fields.Status.Name, issue.Fields.Assignee.DisplayName,
			issue.Fields.Summary)
	}

}

func PrintIssueByKey(jiraKey string) {
	issue, err := gojira.GetIssue(jiraKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Issue: %s\n", issue.Key)
	fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.DisplayName)
	fmt.Printf("Status: %s\n", issue.Fields.Status.Name)
	fmt.Printf("Summary: %s\n\n", issue.Fields.Summary)
	var desc string
	if issue.Fields.Description != nil {
		desc = issue.Fields.Description.(string)
	} else {
		desc = "No description"
	}
	fmt.Printf("Description: \n%s\n", desc)

}

func PrintTransitionsOfIssue(jiraKey string) {
	issue, err := gojira.GetIssue(jiraKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	transitions, err := issue.GetTransitions()
	if err != nil {
		fmt.Println(err)
	}
	for _, transition := range transitions.Transitions {
		fmt.Printf("%3s %s \n", transition.Id, transition.To.Name)
	}

}

func checkActive(issueKey string) bool {
	issueTmpFile := tmpDir + issueKey
	_, err := os.Stat(issueTmpFile)
	if err != nil {
		return false
	} else {
		return true
	}

}

func checkCurrentIssuesInProgress() bool {
	fileInfo, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		panic(err)
	}
	if (len(fileInfo)) > 0 {
		return true
	} else {
		return false
	}

}
func startProgress(issueKey string) error {
	if checkActive(issueKey) {
		return errors.New("Issue already started")
	}
	if checkCurrentIssuesInProgress() {
		return errors.New("You alredy have started issue")
	}
	err := handleHooks("pre_start", issueKey)
	if err != nil {
		return err
	}
	_, err = os.Create(tmpDir + issueKey)

	if err != nil {
		return err
	}
	err = handleHooks("post_start", issueKey)
	if err != nil {
		return err
	}

	return nil
}

func termProgress(issueKey string) error {
	if checkActive(issueKey) {
		err := handleHooks("pre_stop", issueKey)
		if err != nil {
			return err
		}

		fi, err := os.Stat(tmpDir + issueKey)
		if err != nil {
			return err
		}
		dur := time.Now().Sub(fi.ModTime())
		rawMinutes := int(dur.Minutes())
		var wlHours int
		var wlMinutes int
		residue := rawMinutes % 60
		if residue == rawMinutes {
			wlHours = 0
		} else {
			wlMinutes = residue
			wlHours = rawMinutes / 60
		}
		wlTotal := fmt.Sprintf("%dh %dm", wlHours, wlMinutes)
		err = workLog(issueKey, wlTotal)
		if err != nil {
			return err
		}
		err = os.Remove(tmpDir + issueKey)
		err = handleHooks("post_stop", issueKey)
		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("Selected issue not started")
	}
}

func workLog(issueKey, worklogTime string) error {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("You have worked %s\n", worklogTime)
	fmt.Println("Would you like log your work time? (Y)es)/(A)bort/(N)o")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if strings.Trim(strings.ToUpper(text), "\n") == "Y" {
		fmt.Println("Enter description by one line")
		logReader := bufio.NewReader(os.Stdin)
		log, _ := logReader.ReadString('\n')
		err = issue.SetWorklog(worklogTime, log)
		if err != nil {
			return err
		}
		return nil
	}
	if strings.Trim(strings.ToUpper(text), "\n") == "A" {
		return errors.New("Aborting")
	} else {
		fmt.Println("Stop without logging")
		return nil
	}

}

func moveIssue(issueKey, transitionId string) error {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		fmt.Println(err)
	}
	var b = []byte(fmt.Sprintf(`{"transition": {"id": "%s"}}`, transitionId))
	err = issue.SetTransition(bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}

func printComments(issueKey string) {
	issue, err := gojira.GetIssue(issueKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	comments, err := issue.GetComments()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, comment := range comments.Comments {
		fmt.Printf("\n################\n")
		fmt.Printf("ID: %s\n", comment.Id)
		fmt.Printf("Author: %s\n", comment.Author.DisplayName)
		fmt.Printf("Update: %s\n", comment.Updated)
		fmt.Printf("Comment: \n%s\n", comment.Body)
	}

}

func handleHooks(stageName, jiraKey string) error {
	execHooks := func(hookName string) error {
		fmt.Printf("Execute hook %s\n", hookName)
		err := exec.Command(hookName, jiraKey, config.exportToHook()).Run()
		if err != nil {
			return errors.New(fmt.Sprintf("Hook %s failed\n", hookName))
		}
		fmt.Printf("Hook %s successfully ended\n", hookName)
		return nil
	}
	switch stageName {
	case "pre_start":
		for _, hookName := range config.Hooks.PreStart {
			execHooks(hookName)
		}
	case "post_start":
		for _, hookName := range config.Hooks.PostStart {
			execHooks(hookName)
		}
	case "pre_stop":
		for _, hookName := range config.Hooks.PreStop {
			execHooks(hookName)
		}
	case "post_stop":
		for _, hookName := range config.Hooks.PostStop {
			execHooks(hookName)
		}
	default:
		return errors.New("Unknown hooks stage")

	}
	return nil
}
