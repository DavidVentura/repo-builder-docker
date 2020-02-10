package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

type GogsPushEvent struct {
	Ref        string `json:"ref"`
	Ref_type   string `json:"ref_type"`
	Repository struct {
		Ssh_url string `json:"ssh_url"`
	} `json:"repository"`
}

type Configuration struct {
	Repos []Repo
}

type HookData struct {
	GitUrl string
	Tag    string
}
type Repo struct {
	Name                      string
	GitUrl                    string
	RelativePathForDockerfile string
	dstPath                   string
}

const repoCloneBase = "/tmp/"

var config Configuration

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

func notification(msg string, output io.Writer) error {

	output.Write([]byte("Sending telegram notification..\n"))
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TELEGRAM_BOT_KEY"))

	j, err := json.Marshal(map[string]string{"chat_id": os.Getenv("TELEGRAM_CHAT_ID"),
		"text": msg})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Write([]byte("Failure sending telegram notification..\n"))
		output.Write([]byte(err.Error()))
		return err
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	//_, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
	output.Write([]byte("Succesfully sent telegram notification..\n"))
	return nil
}

func dockerBuild(repo Repo, hookData HookData, output io.Writer) error {
	dockerfileDir := path.Dir(path.Join(repo.dstPath, repo.RelativePathForDockerfile))

	buildCmd := exec.Command("docker", "build",
		"--build-arg", fmt.Sprintf("TAG=%s", hookData.Tag),
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

func on_hook(repo Repo, hookData HookData) {
	repo.dstPath = path.Join(repoCloneBase, repo.Name)
	buildLog, err := os.Create(fmt.Sprintf("/tmp/%s.log", repo.Name))

	if err != nil {
		panic(err)
	}
	defer buildLog.Close()
	defer buildLog.Sync()

	mw := io.MultiWriter(buildLog, os.Stdout)
	err = cloneRepo(repo, mw)
	if err != nil {
		mw.Write([]byte("Failed to clone repo!\n"))
		os.Exit(1)
	}
	notification(fmt.Sprintf("Starting build for %s", repo.Name), mw)
	err = dockerBuild(repo, hookData, mw)
	if err != nil {
		notification("Build failed!", mw)
		mw.Write([]byte("Failed to build repo!\n"))
		os.Exit(1)
	}
	notification("Build succeeded!", mw)
}

func readConf() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
func hookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header["X-Gogs-Event"][0] != "create" {
		fmt.Println("Event was not 'create', refusing to work")
		fmt.Println(r.Header)
		return
	}
	var event GogsPushEvent
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &event)
	fmt.Println(event)
	if event.Ref_type != "tag" {
		fmt.Println("Received an event that is not from a tag. Aborting")
		return
	}
	for _, repo := range config.Repos {
		if repo.GitUrl == event.Repository.Ssh_url {
			fmt.Println("This repo matches!")
			hookData := HookData{
				GitUrl: repo.GitUrl,
				Tag:    event.Ref,
			}
			go on_hook(repo, hookData)
		}
	}
}

func hookEndpoint() {
	http.HandleFunc("/hook", hookHandler)
	http.ListenAndServe(":8080", nil)
}

func main() {
	readConf()
	fmt.Println(config)
	hookEndpoint()
}
