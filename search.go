package main

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/tears-of-noobs/gojira"
)

func getIssues(
	projectName string,
	query string,
	limit int,
) (*gojira.JiraSearchIssues, error) {
	params := []string{
		"project = " + projectName,
	}

	if query != "" {
		params = append(params, "("+query+")")
	}

	sort := "ORDER BY updated DESC"

	jql := strings.Join(params, " AND ") + " " + sort

	request := url.QueryEscape(jql) +
		"&fields=key,summary,status,assignee&maxResults=" + strconv.Itoa(limit)

	reply, err := gojira.RawSearch(request)
	if err != nil {
		return nil, err
	}

	var result gojira.JiraSearchIssues
	err = json.Unmarshal(reply, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
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
