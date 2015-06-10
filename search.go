package main

import (
	"encoding/json"
	"strconv"

	"github.com/tears-of-noobs/gojira"
)

func searchIssues(
	username string,
	projectName string,
	issueListCount int,
) (*gojira.JiraSearchIssues, error) {
	searchQuery := "project%20%3D%20" + projectName +
		"%20AND%20assignee%20%3D%20" + username + "%20order%20by%20updated%20DESC" +
		"&fields=key,summary,status,assignee&maxResults=" + strconv.Itoa(issueListCount)

	jsonedSearchIssues, err := gojira.RawSearch(searchQuery)
	if err != nil {
		return nil, err
	}

	var searchIssues gojira.JiraSearchIssues
	err = json.Unmarshal(jsonedSearchIssues, &searchIssues)
	if err != nil {
		return nil, err
	}

	return &searchIssues, nil
}

func searchIssuesByFilterID(
	filterID int,
) (*gojira.JiraSearchIssues, error) {
	jsonedSearchIssues, err := gojira.FilterSearch(filterID)
	if err != nil {
		return nil, err
	}

	var searchIssues gojira.JiraSearchIssues
	err = json.Unmarshal(jsonedSearchIssues, &searchIssues)
	if err != nil {
		return nil, err
	}

	return &searchIssues, nil
}
