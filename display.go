package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/reconquest/karma-go"
	"github.com/reconquest/loreley"
	"github.com/seletskiy/tplutil"
	"github.com/tears-of-noobs/gojira"
)

var DefaultTemplate = template.Must(
	template.New("default").Parse(
		"{{.mark}}{{.key}}\t{{.stage}}\t{{.name}}\t{{.summary}}",
	),
)

var OnlySummaryTemplate = template.Must(
	template.New("only-summary").Parse(
		"{{.summary}}",
	),
)

func displayIssues(
	issues []gojira.Issue,
	activeIssueKey string,
	showName bool,
	onlySummary bool,
	workflow Workflow,
) error {
	var err error

	buffer := bytes.NewBuffer(nil)
	board := tabwriter.NewWriter(buffer, 1, 4, 2, ' ', 0)

	templates := map[string]*template.Template{}

	for _, issue := range issues {
		issueMark := ""
		isActive := false
		if issue.Key == activeIssueKey {
			issueMark = "* "
			isActive = true
		}

		name := issue.Fields.Assignee.Name
		if !showName {
			name = issue.Fields.Assignee.DisplayName
		}

		view := map[string]interface{}{
			"is_active":             isActive,
			"mark":                  issueMark,
			"key":                   issue.Key,
			"stage":                 issue.Fields.Status.Name,
			"name":                  name,
			"assignee_name":         issue.Fields.Assignee.Name,
			"assignee_display_name": issue.Fields.Assignee.DisplayName,
			"summary":               issue.Fields.Summary,
		}

		tpl := DefaultTemplate
		if onlySummary {
			tpl = OnlySummaryTemplate
		} else {
			if stage, ok := workflow.GetStage(issue.Fields.Status.Name); ok {
				// skip the issue if the stage's order is -1
				if stage.Order == -1 {
					continue
				}
				if stage.Template != "" {
					tpl, ok = templates[issue.Fields.Status.Name]
					if !ok {
						tpl = template.New(issue.Fields.Status.Name)
						tpl, err = tpl.Parse(stage.Template)
						if err != nil {
							return karma.Format(
								err,
								"unable to parse template: %s",
								issue.Fields.Status.Name,
							)
						}
					}
				}
			}
		}

		contents, err := tplutil.ExecuteToString(tpl, view)
		if err != nil {
			return karma.Format(
				err,
				"unable to execute template: %s", tpl.Name(),
			)
		}

		board.Write([]byte(contents + "\n"))
	}

	board.Flush()

	loreley.DelimLeft = "<"
	loreley.DelimRight = ">"

	result, err := loreley.CompileAndExecuteToString(
		strings.NewReplacer("<", `<"<">`).Replace(buffer.String()),
		nil,
		nil,
	)
	if err != nil {
		return karma.Format(
			err,
			"unable to colorize output",
		)
	}

	fmt.Print(result)

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
