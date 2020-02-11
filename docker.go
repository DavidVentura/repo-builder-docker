package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

func dockerBuild(repo Repo, hookData HookData, output io.Writer) error {
	dockerfileDir := path.Dir(path.Join(repo.dstPath, repo.RelativePathForDockerfile))

	buildCmd := exec.Command("docker", "build",
		"--build-arg", fmt.Sprintf("TAG=%s", hookData.Tag),
		"--build-arg", fmt.Sprintf("BUCKET_NAME=%s", repo.Bucket),
		"--build-arg", fmt.Sprintf("S3_ACCESS_KEY=%s", os.Getenv("S3_ACCESS_KEY")),
		"--build-arg", fmt.Sprintf("S3_SECRET_KEY=%s", os.Getenv("S3_SECRET_KEY")),
		dockerfileDir,
	)
	output.Write([]byte(fmt.Sprintf("Running command: %s\n", buildCmd.String())))

	buildCmd.Stdout = output
	buildCmd.Stderr = output
	err := buildCmd.Run()
	if err != nil {
		output.Write([]byte("Execution Failed!!"))
		return err
	} else {
		output.Write([]byte("Execution successful!!"))
	}
	return nil
}
