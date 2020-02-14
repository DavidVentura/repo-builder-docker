package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

var artifactUploaderDockerfile = `
FROM builder
RUN apt-get install --no-install-recommends -y s4cmd curl
ARG TAG
ARG S3_ACCESS_KEY
ARG S3_SECRET_KEY
ARG BUCKET_NAME
ARG REPO_NAME
ENV http_proxy=
RUN s4cmd --endpoint-url=http://ci.labs:9000 ls s3://${BUCKET_NAME}/${TAG}/ || s4cmd --endpoint-url=http://ci.labs:9000 mb s3://${BUCKET_NAME}/${TAG}/
RUN while read -r artifact; do s4cmd --endpoint-url=http://ci.labs:9000 put --force $artifact s3://${BUCKET_NAME}/${TAG}/; done </artifacts
RUN curl -s http://david-dotopc:8080/deploy/${REPO_NAME}/${TAG}
`

func dockerBuild(repo Repo, hookData HookData, output io.Writer) error {
	dockerfile := path.Join(repo.dstPath, repo.RelativePathForDockerfile)
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		return err
	}

	fd, err := os.Open(dockerfile)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	extendedContent := []byte(string(content) + artifactUploaderDockerfile)

	wfd, err := os.Create(dockerfile)
	if err != nil {
		return err
	}
	wfd.Write(extendedContent)
	wfd.Close()

	buildCmd := exec.Command("docker", "build",
		"--build-arg", fmt.Sprintf("TAG=%s", hookData.Tag),
		"--build-arg", fmt.Sprintf("REPO_NAME=%s", repo.Name),
		"--build-arg", fmt.Sprintf("BUCKET_NAME=%s", repo.Bucket),
		"--build-arg", fmt.Sprintf("S3_ACCESS_KEY=%s", os.Getenv("S3_ACCESS_KEY")),
		"--build-arg", fmt.Sprintf("S3_SECRET_KEY=%s", os.Getenv("S3_SECRET_KEY")),
		path.Dir(dockerfile),
	)

	output.Write([]byte(fmt.Sprintf("Running command: %s\n", buildCmd.String())))

	buildCmd.Stdout = output
	buildCmd.Stderr = output
	return buildCmd.Run()
}
