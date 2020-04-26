package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

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
		"--build-arg", fmt.Sprintf("TAG=%s", hookData.Ref),
		"--build-arg", fmt.Sprintf("REPO_NAME=%s", repo.Name),
		"--build-arg", fmt.Sprintf("SUBPROJECT=%s", subproject.Name),
		"--build-arg", fmt.Sprintf("BUCKET_NAME=%s", repo.Bucket),
		"--build-arg", fmt.Sprintf("ARTIFACTS=%s", strings.Join(subproject.Artifacts, "\n")),

		"--build-arg", fmt.Sprintf("S3_ACCESS_KEY=%s", os.Getenv("S3_ACCESS_KEY")),
		"--build-arg", fmt.Sprintf("S3_SECRET_KEY=%s", os.Getenv("S3_SECRET_KEY")),
		subprojectDir,
	)

	output.Write([]byte(fmt.Sprintf("Running command: %s\n", buildCmd.String())))

	buildCmd.Stdout = output
	buildCmd.Stderr = output
	return buildCmd.Run()
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
