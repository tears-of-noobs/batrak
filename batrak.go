package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tears-of-noobs/gojira"
)

func PrintIssues(user string) {
	searchString := "project%20%3D%20" + projectName +
		"%20AND%20assignee%20%3D%20" + user + "%20order%20by%20key%20DESC" +
		"&fields=key,summary,status&maxResults=10"
	result, err := gojira.RawSearch(searchString)
	if err != nil {
		fmt.Println(err)
	}
	var jiraIssues gojira.JiraSearchIssues
	err = json.Unmarshal(result, &jiraIssues)
	if err != nil {
		fmt.Println(err)
	}

	for _, issue := range jiraIssues.Issues {
		var started string
		if checkActive(issue.Key) {
			started = "*"
		}
		fmt.Printf("%2s %10s %13s %s\n", started, issue.Key,
			issue.Fields.Status.Name, issue.Fields.Summary)
	}

}

func PrintIssueByKey(jiraKey string) {
	issue, err := gojira.GetIssue(jiraKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Задача: %s\n", issue.Key)
	fmt.Printf("Статус: %s\n", issue.Fields.Status.Name)
	fmt.Printf("Название: %s\n\n", issue.Fields.Summary)
	var desc string
	if issue.Fields.Description != nil {
		desc = issue.Fields.Description.(string)
	} else {
		desc = "Нет описания"
	}
	fmt.Printf("Описание: \n%s\n", desc)

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
	_, err := os.Create(tmpDir + issueKey)
	if err != nil {
		return err
	}
	return nil
}

func termProgress(issueKey string) error {
	if checkActive(issueKey) {
		fi, err := os.Stat(tmpDir + issueKey)
		if err != nil {
			return err
		}
		dur := time.Now().Sub(fi.ModTime())
		wlHours := strconv.FormatFloat(dur.Hours(), 'f', 0, 64)
		wlMinutes := strconv.FormatFloat(dur.Minutes(), 'f', 0, 64)
		wlTotal := fmt.Sprintf("%sh %sm", wlHours, wlMinutes)
		err = os.Remove(tmpDir + issueKey)
		workLog(issueKey, wlTotal)
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
	fmt.Println("Would you like log your work time?")
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
		fmt.Printf("Author: %s\n", comment.Author.DisplayName)
		fmt.Printf("Update: %s\n", comment.Updated)
		fmt.Printf("Comment: \n%s\n", comment.Body)
	}

}
