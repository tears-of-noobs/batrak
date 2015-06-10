package main

import (
	"bytes"
	"encoding/json"

	"github.com/tears-of-noobs/gojira"
)

func addComment(issue *gojira.Issue) error {
	comment, err := editTemporaryFile(issue.Key + "-batrak-issue-comment")
	if err != nil {
		return err
	}

	jsonedRequest, err := json.Marshal(map[string]string{"body": comment})
	if err != nil {
		return err
	}

	_, err = issue.SetComment(bytes.NewBuffer(jsonedRequest))
	if err != nil {
		return err
	}

	return nil
}
