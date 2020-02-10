package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func cloneRepo(repo Repo, output io.Writer) error {
	output.Write([]byte(fmt.Sprintf("Requested to clone %s\n", repo)))
	if _, err := os.Stat(repo.dstPath); os.IsNotExist(err) {
		cloneCmd := exec.Command("git", "clone", repo.GitUrl, repo.dstPath)
		out, err := cloneCmd.CombinedOutput()
		output.Write([]byte(fmt.Sprintf("Output of git clone of %s", repo)))
		output.Write([]byte(out))
		if err != nil {
			output.Write([]byte(fmt.Sprintf("Git clone of %s failed: %s\n", repo, err)))
			return err
		}
	} else {
		output.Write([]byte("No need to clone the repository as it exists already\n"))
	}
	return nil
}
