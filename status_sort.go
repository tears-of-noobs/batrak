package main

import (
	"fmt"
	"sort"

	"github.com/tears-of-noobs/gojira"
)

func getWorkflowStageStatusOrder(issue gojira.Issue, workflowStages []Stage) int {
	if len(workflowStages) == 0 {
		return 1
	}
	for _, stage := range workflowStages {
		if stage.Name == issue.Fields.Status.Name {
			return stage.Order
		}
	}

	fmt.Println("Unknown workflow stage:", issue.Fields.Status.Name)
	return -1
}

type StatusSortableIssues struct {
	Issues []gojira.Issue
	Stages []Stage
}

func (sortable StatusSortableIssues) Len() int {
	return len(sortable.Issues)
}

func (sortable StatusSortableIssues) Swap(i, j int) {
	sortable.Issues[i], sortable.Issues[j] =
		sortable.Issues[j], sortable.Issues[i]
}

func (sortable StatusSortableIssues) Less(i, j int) bool {
	return getWorkflowStageStatusOrder(sortable.Issues[i], sortable.Stages) <
		getWorkflowStageStatusOrder(sortable.Issues[j], sortable.Stages)
}

func sortIssuesByStatus(issues []gojira.Issue, stages []Stage) []gojira.Issue {
	sortable := StatusSortableIssues{
		Issues: issues,
		Stages: stages,
	}

	sort.Sort(sortable)

	return sortable.Issues
}
