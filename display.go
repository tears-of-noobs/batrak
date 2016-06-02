package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/tears-of-noobs/gojira"
)

func displayIssues(issues []gojira.Issue, activeIssueKey string) error {
	board := tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)

	for _, issue := range issues {
		issueMark := ""
		if issue.Key == activeIssueKey {
			issueMark = "* "
		}

		fmt.Fprintf(
			board,
			"%s%s\t%s\t%s\t%s\n",
			issueMark,
			issue.Key,
			issue.Fields.Status.Name,
			issue.Fields.Assignee.DisplayName,
			issue.Fields.Summary,
		)
	}

	board.Flush()

	return nil
}

func displayIssue(issue *gojira.Issue) error {
	fmt.Printf("Issue:    %s\n", issue.Key)
	fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.DisplayName)
	fmt.Printf("Status:   %s\n", issue.Fields.Status.Name)
	fmt.Printf("Summary:  %s\n\n", issue.Fields.Summary)

	desc := "<no description>"
	if issue.Fields.Description != nil {
		desc = issue.Fields.Description.(string)
	}

	fmt.Printf("%s\n", desc)

	return nil
}

func displayTransitions(transitions *gojira.Transitions) error {
	for _, transition := range transitions.Transitions {
		fmt.Printf("%3s %s \n", transition.Id, transition.To.Name)
	}

	return nil
}

func displayComments(comments *gojira.Comments) error {
	for _, comment := range comments.Comments {
		fmt.Printf("\n################\n")
		fmt.Printf("ID:     %s\n", comment.Id)
		fmt.Printf("Author: %s\n", comment.Author.DisplayName)
		fmt.Printf("Update: %s\n", comment.Updated)
		fmt.Printf("Comment: \n%s\n", comment.Body)
	}

	return nil
}
