package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/reconquest/executil-go"
	"github.com/reconquest/karma-go"
	"github.com/tears-of-noobs/gojira"
	"github.com/zazab/zhash"
)

const (
	prefaceCreateIssue = `

# Write a summary & description for this issue.
# The first line of text is the summary and the rest is description.
`
)

func handleCreateMode(project, issueType string) error {
	contents, err := editTemporaryFile(prefaceCreateIssue, ".batrak")
	if err != nil {
		if executil.IsExitError(err) {
			return nil
		}

		return err
	}

	var summary, desc []string
	for _, line := range strings.Split(contents, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}

		if line == "" && len(summary) == 0 {
			summary = desc[:]
			desc = []string{}
			continue
		}

		desc = append(desc, line)
	}

	if len(summary) == 0 {
		summary = desc[:]
		desc = []string{}
	}

	if len(summary) == 0 && len(desc) == 0 {
		log.Println("Aborted")
		return nil
	}

	fields := map[string]string{}
	i := len(desc) - 1
	for ; i > 0; i-- {
		if desc[i] == "" {
			if i == len(desc)-1 {
				continue
			}

			break
		}
		if !strings.HasPrefix(desc[i], "$") {
			break
		}
		if !strings.Contains(desc[i], ":") {
			break
		}
		chunks := strings.SplitN(desc[i], ":", 2)

		fields[chunks[0][1:]] = chunks[1]
	}

	desc = desc[:i+1]

	hash := zhash.NewHash()
	hash.Set(project, "fields", "project", "key")
	hash.Set(strings.Join(summary, "\n"), "fields", "summary")
	hash.Set(strings.Join(desc, "\n"), "fields", "description")
	hash.Set(issueType, "fields", "issuetype", "name")
	for key, value := range fields {
		hash.Set(value, append([]string{"fields"}, strings.Split(key, ".")...)...)
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(hash.GetRoot())
	if err != nil {
		return err
	}

	issue, err := gojira.CreateIssue(buffer)
	if err != nil {
		return karma.Format(
			err,
			"unable to create issue",
		)
	}

	fmt.Println(issue.Key)

	return nil
}
