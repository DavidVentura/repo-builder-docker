package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func cloneRepo(repo Repo, hookData HookData, output io.Writer) error {
	output.Write([]byte(fmt.Sprintf("Requested to clone %s\n", repo.Name)))
	commands := [][]string{
		{"clean", "-fd"},
		{"reset", "--hard"},
		{"checkout", "master"},
		{"pull"},
		{"checkout", hookData.Ref},
	}
	if _, err := os.Stat(repo.dstPath); os.IsNotExist(err) {
		commands = append([][]string{{"clone", repo.GitUrl, repo.dstPath}}, commands...)
		err = os.Mkdir(repo.dstPath, 0700)
		if err != nil {
			delerr := os.Remove(repo.dstPath)
			if delerr != nil {
				output.Write([]byte("Failed to delete the empty path:\n"))
				output.Write([]byte(delerr.Error()))
			}
			output.Write([]byte("Failed to clone the repo:\n"))
			output.Write([]byte(err.Error()))
			return err
		}
	} else {
		output.Write([]byte("No need to clone the repository as it exists already\n"))
	}

	for _, command := range commands {
		output.Write([]byte(strings.Join(command, " ")))
		cmd := exec.Command("git", command...)
		cmd.Dir = repo.dstPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		output.Write([]byte(fmt.Sprintf("Output of %s of %s\n", cmd.String(), repo)))
		output.Write([]byte(out))
	}
	return nil
}
