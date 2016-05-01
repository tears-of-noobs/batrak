package main

import "github.com/tears-of-noobs/gojira"

func addComment(issue *gojira.Issue) error {
	comment, err := editTemporaryFile(issue.Key + "-batrak-issue-comment")
	if err != nil {
		return err
	}

	_, err = issue.SetComment(&gojira.Comment{Body: comment})
	if err != nil {
		return err
	}

	return nil
}
