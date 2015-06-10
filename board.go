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
) (kanbanBoard, error) {
	if len(workflowStages) == 0 {
		return kanbanBoard{}, fmt.Errorf("kanban stages is not defined")
	}

	board := kanbanBoard{
		issues: issues,
		stages: workflowStages,
	}

	return board, nil
}

func (board *kanbanBoard) GenerateBoardData(activeIssueKey string) {
	board.tableHeaders = make([]string, len(board.stages))
	for headerIndex, stage := range board.stages {
		board.tableHeaders[headerIndex] = stage.Name
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

			board.tableRows[rowIndex][headerIndex] = stageIssuesMap[stage][0].Key

			if board.tableRows[rowIndex][headerIndex] == activeIssueKey {
				board.tableRows[rowIndex][headerIndex] = "*" +
					board.tableRows[rowIndex][headerIndex]
			}

			stageIssuesMap[stage] = stageIssuesMap[stage][1:]
		}
	}
}
func (board kanbanBoard) Display() {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader(board.tableHeaders)
	for _, v := range board.tableRows {
		table.Append(v)
	}

	table.Render()
}
