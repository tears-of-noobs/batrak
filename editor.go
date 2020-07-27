package main

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func editTemporaryFile(preface, suffix string) (string, error) {
	temporaryFile, err := ioutil.TempFile(os.TempDir(), "*"+suffix)
	if err != nil {
		return "", err
	}

	if preface != "" {
		_, err = temporaryFile.WriteString(preface)
		if err != nil {
			return "", err
		}

		temporaryFile.Seek(0, 0)
	}

	defer temporaryFile.Close()

	cmd := exec.Command(os.Getenv("EDITOR"), temporaryFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(temporaryFile)
	if err != nil {
		return "", err
	}

	err = os.Remove(temporaryFile.Name())
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
