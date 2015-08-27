package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tears-of-noobs/gojira"
)

func startProgress(issueKey string, hooks Hooks) error {
	err := hooks.Handle("pre_start", issueKey)
	if err != nil {
		return err
	}

	err = setActiveIssueKey(issueKey)
	if err != nil {
		return err
	}

	err = hooks.Handle("post_start", issueKey)
	if err != nil {
		return err
	}

	return nil
}

func stopProgress(issue *gojira.Issue, hooks Hooks) error {
	err := hooks.Handle("pre_stop", issue.Key)
	if err != nil {
		return err
	}

	err = hooks.Handle("pre_stop", issue.Key)
	if err != nil {
		return err
	}

	hours, minutes, err := getActiveIssueTime()
	if err != nil {
		return err
	}

	loggingTime := fmt.Sprintf("%dh %dm", hours, minutes)

	fmt.Printf("You have worked %s\n", loggingTime)

	for {
		fmt.Println("Do you want to log this time? (Y)es)/(A)bort/(N)o")

		userAnswer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}

		stopPromptLoop := false

		switch strings.Trim(strings.ToUpper(userAnswer), "\n") {
		case "Y":
			worklogMessage, err := editTemporaryFile(
				issue.Key + "-worklog-message.tmp",
			)
			if err != nil {
				return err
			}

			err = issue.SetWorklog(loggingTime, string(worklogMessage))
			if err != nil {
				return err
			}

			fmt.Println("Issue progress stopped")
			stopPromptLoop = true

		case "N":
			err = issue.SetWorklog(loggingTime, "")
			if err != nil {
				return err
			}

			fmt.Println("Issue progress stopped without logging")
			stopPromptLoop = true

		case "A":
			return nil
		}

		if stopPromptLoop {
			break
		}
	}

	err = setActiveIssueKey("")
	if err != nil {
		return err
	}

	err = hooks.Handle("post_stop", issue.Key)
	if err != nil {
		return err
	}

	fmt.Printf("Issue %s stopped\n", issue.Key)

	return nil
}

func getActiveIssueKey() (string, error) {
	filename, err := getActiveIssueFilename()
	if err != nil {
		return "", err
	}

	issueKey, err := ioutil.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	return string(issueKey), nil
}

func setActiveIssueKey(issueKey string) error {
	filename, err := getActiveIssueFilename()
	if err != nil {
		return err
	}

	if issueKey == "" {
		return os.Remove(filename)
	}

	return ioutil.WriteFile(filename, []byte(issueKey), 0700)
}

func getActiveIssueTime() (hours int, minutes int, err error) {
	filename, err := getActiveIssueFilename()
	if err != nil {
		return 0, 0, err
	}

	fileinfo, err := os.Stat(filename)
	if err != nil {
		return 0, 0, err
	}

	modifyDate := time.Now().Sub(fileinfo.ModTime())
	totalMinutes := int(modifyDate.Minutes())

	hours = totalMinutes / 60
	if hours == 0 {
		minutes = totalMinutes
	} else {
		minutes = totalMinutes % 60
	}

	return hours, minutes, nil
}

func getActiveIssueFilename() (string, error) {
	batrakDirectory := filepath.Join(os.Getenv("HOME"), "/.batrak/")
	activeIssueFilename := filepath.Join(batrakDirectory, "active-issue")

	_, err := os.Stat(batrakDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", nil
		}

		err = os.Mkdir(batrakDirectory, 0700)
		if err != nil {
			return "", err
		}

		return activeIssueFilename, nil
	}

	return activeIssueFilename, nil
}
