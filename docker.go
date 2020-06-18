package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type FilePair struct {
	artifact   string
	outputPath string
}

func dockerBuild(repo Repo,
	hookData HookData,
	output io.Writer,
	subproject SubProjectConfig) error {
	/*
		if _, err := os.Stat(repo.dstPath); os.IsNotExist(err) {
			return err
		}

		fd, err := os.Open(config.BuildDockerfilePath)
		if err != nil {
			return err
		}
	*/

	subprojectDir := path.Join(repo.dstPath, subproject.Dir)
	dockerfileDest := path.Join(subprojectDir, "Dockerfile")

	err := fileCopy(config.BuildDockerfilePath, dockerfileDest)
	if err != nil {
		return err
	}
	output.Write([]byte(fmt.Sprintf("Copied %s to %s\n", config.BuildDockerfilePath, dockerfileDest)))

	buildCmd := exec.Command("docker", "build",
		"-t", strings.ToLower(fmt.Sprintf("%s-%s:%s", repo.Name, subproject.Name, hookData.Ref)),
		subprojectDir,
	)

	output.Write([]byte(fmt.Sprintf("Running command: %s\n", buildCmd.String())))

	buildCmd.Stdout = output
	buildCmd.Stderr = output
	return buildCmd.Run()
}
func dockerCopyFile(repo Repo,
	hookData HookData,
	output io.Writer,
	subproject SubProjectConfig,
	fp FilePair) error {

	imageName := strings.ToLower(fmt.Sprintf("%s-%s:%s", repo.Name, subproject.Name, hookData.Ref))
	createCmd := exec.Command("docker", "create", "-ti", "--name", "dummy", imageName, "bash")
	output.Write([]byte(strings.Join(createCmd.Args, " ")))
	output.Write([]byte(fmt.Sprintf("Output..\n")))
	createCmd.Stdout = output
	createCmd.Stderr = output
	createCmd.Run()

	cpCmd := exec.Command("docker", "cp", fmt.Sprintf("dummy:/usr/src/app/%s", fp.artifact), fp.outputPath)
	output.Write([]byte(strings.Join(cpCmd.Args, " ")))
	cpCmd.Stdout = output
	cpCmd.Stderr = output
	cpCmd.Run()

	deleteCmd := exec.Command("docker", "rm", "-f", "dummy")
	output.Write([]byte(strings.Join(deleteCmd.Args, " ")))
	deleteCmd.Stdout = output
	deleteCmd.Stderr = output
	deleteCmd.Run()
	return nil
}

func fileCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
