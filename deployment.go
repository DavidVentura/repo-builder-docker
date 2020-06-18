package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func BuildUploadAndDeploy(repo Repo,
	hookData HookData,
	buildLog io.Writer,
	subproject SubProjectConfig) error {

	err := dockerBuild(repo, hookData, buildLog, subproject)
	if err != nil {
		Log.Printf("Failed building repo %+v", repo)
		notifications <- Notification{msg: "Build failed!", chat_id: repo.TelegramChatId}
		buildLog.Write([]byte(fmt.Sprintf("Failed to build %s.%s!\n%s\n", repo.Name, subproject.Name, err.Error())))
		return err
	}
	err = uploadDockerArtifact(repo, hookData, buildLog, subproject)
	if err != nil {
		Log.Printf("Failed uploading build artifacts from repo %+v", repo)
		notifications <- Notification{msg: "Failed to upload artifacts!", chat_id: repo.TelegramChatId}
		buildLog.Write([]byte(fmt.Sprintf("Failed upload artifact for build %s.%s!\n%s\n", repo.Name, subproject.Name, err.Error())))
		return err
	}
	err = deployRepo(repo, hookData, buildLog, subproject)
	if err != nil {
		Log.Printf("Failed deploying repo %+v", repo)
		notifications <- Notification{msg: "Failed to deploy build!", chat_id: repo.TelegramChatId}
		buildLog.Write([]byte(fmt.Sprintf("Failed deploying build %s.%s!\n%s\n", repo.Name, subproject.Name, err.Error())))
		return err
	}
	return nil
}

func deployRepo(repo Repo,
	hookData HookData,
	buildLog io.Writer,
	subproject SubProjectConfig) error {
	// FIXME
	url := fmt.Sprintf("%s/%s/%s/%s", config.DeploymentBaseUri, repo.Name, subproject.Name, hookData.Ref)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	buildLog.Write(body)
	return nil
}
func uploadDockerArtifact(repo Repo,
	hookData HookData,
	buildLog io.Writer,
	subproject SubProjectConfig) error {

	for _, artifact := range subproject.Artifacts {
		fd, err := ioutil.TempFile("/tmp/", "tmpDockerFileArtifact") // FIXME uniq?
		if err != nil {
			return err
		}
		tempName := fd.Name()
		defer os.Remove(tempName)
		fd.Close()

		fp := FilePair{artifact: artifact, outputPath: tempName}
		err = dockerCopyFile(repo, hookData, buildLog, subproject, fp)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s/%s", subproject.Name, hookData.Ref, artifact)
		buildLog.Write([]byte(fmt.Sprintf("Uploading to S3 %s:%s", repo.Bucket, key)))
		err = UploadS3(tempName, repo.Bucket, key)
		if err != nil {
			return err
		}
	}

	return nil
}
