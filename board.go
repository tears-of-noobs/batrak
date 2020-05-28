package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/tears-of-noobs/gojira"
)

type kanbanBoard struct {
	issues       []gojira.Issue
	stages       []Stage
	tableRows    [][]string
	tableHeaders []string
	showSummary  bool
	showName     bool
}

type KanbanOrderSortableStages []Stage

func (sortable KanbanOrderSortableStages) Len() int {
	return len(sortable)
}

func (sortable KanbanOrderSortableStages) Swap(i, j int) {
	sortable[i], sortable[j] = sortable[j], sortable[i]
}

func (sortable KanbanOrderSortableStages) Less(i, j int) bool {
	return sortable[i].KanbanOrder < sortable[j].KanbanOrder
}

func NewKanbanBoard(
	issues []gojira.Issue,
	workflowStages []Stage,
	showSummary bool,
	showName bool,
) (kanbanBoard, error) {
	if len(workflowStages) == 0 {
		return kanbanBoard{}, fmt.Errorf("kanban stages are not defined")
	}

	board := kanbanBoard{
		issues:      issues,
		stages:      workflowStages,
		showSummary: showSummary,
		showName:    showName,
	}

	return board, nil
}

func (board *kanbanBoard) GenerateBoardData(activeIssueKey string) {
	board.tableHeaders = []string{}
	for _, stage := range board.stages {
		if stage.KanbanOrder != 0 {
			board.tableHeaders = append(board.tableHeaders, stage.Name)
		}
	}

	stageIssuesMap := map[string][]gojira.Issue{}

	for _, issue := range board.issues {
		stage := issue.Fields.Status.Name
		if _, ok := stageIssuesMap[stage]; !ok {
			stageIssuesMap[stage] = []gojira.Issue{}
		}

		stageIssuesMap[stage] = append(stageIssuesMap[stage], issue)
	}

	board.tableRows = [][]string{}
	more := true
	for rowIndex := 0; more; rowIndex++ {
		more = false
		board.tableRows = append(
			board.tableRows,
			make([]string, len(board.tableHeaders)),
		)

		for headerIndex, stage := range board.tableHeaders {
			if len(stageIssuesMap[stage]) == 0 {
				board.tableRows[rowIndex][headerIndex] = ""
				continue
			}

			if len(stageIssuesMap[stage]) > 1 {
				more = true
			}

			issue := stageIssuesMap[stage][0]

			item := issue.Key

			if issue.Key == activeIssueKey {
				item = "*" + item
			}

			if board.showSummary {
				item += " " + issue.Fields.Summary
			}

			if board.showName {
				item += " (" + issue.Fields.Assignee.Name + ")"
			}

			board.tableRows[rowIndex][headerIndex] = item

			stageIssuesMap[stage] = stageIssuesMap[stage][1:]
		}
	}
}

func (board kanbanBoard) Display() {
	table := tablewriter.NewWriter(os.Stdout)
	if board.showSummary {
		table.SetRowLine(true)
		table.SetRowSeparator("-")
	}

	table.SetHeader(board.tableHeaders)
	for _, v := range board.tableRows {
		table.Append(v)
	}

	table.Render()
}
