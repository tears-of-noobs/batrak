package main

import (
	"encoding/json"
	"fmt"
	"gojira"
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
		fmt.Printf("%15s %13s %s\n", issue.Key,
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
	}

	transitions, err := issue.GetTransitions()
	if err != nil {
		fmt.Println(err)
	}
	for _, transition := range transitions.Transitions {
		fmt.Printf("%3s %s \n", transition.Id, transition.To.Name)
	}

}
