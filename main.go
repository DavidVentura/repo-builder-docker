package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type GogsPushEvent struct {
	Ref        string `json:"ref"`
	Ref_type   string `json:"ref_type"`
	Repository struct {
		Ssh_url string `json:"ssh_url"`
	} `json:"repository"`
}

type Configuration struct {
	LogPath             string
	RepoCloneBase       string
	Repos               []Repo
	BuildDockerfilePath string
}

type HookData struct {
	GitUrl string
	Ref    string
}

type RepoBuildConfig struct {
	Subprojects []SubProjectConfig `json:"subprojects"`
}

type SubProjectConfig struct {
	Name      string   `json:"name"`
	Dir       string   `json:"dir"`
	Artifacts []string `json:"artifacts"`
}

type Repo struct {
	Name           string
	GitUrl         string
	Bucket         string
	TelegramChatId int

	dstPath string
}

var config Configuration
var Log *log.Logger
var notifications = make(chan Notification, 2)

func readConf(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if config.RepoCloneBase == "" {
		fmt.Println("repoCloneBase must be provided")
		os.Exit(1)
	}
	if config.LogPath == "" {
		fmt.Println("LogPath must be provided")
		os.Exit(1)
	}
	if config.BuildDockerfilePath == "" {
		fmt.Println("BuildDockerfilePath must be provided")
		os.Exit(1)
	}

	for _, repo := range config.Repos {
		if repo.TelegramChatId == 0 {
			fmt.Printf("For repo %+v, telegram chat id is 0, so no notifications will be sent\n", repo)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println(fmt.Sprintf("Usage: %s config.json", os.Args[0]))
		os.Exit(1)
	}

	readConf(os.Args[1])
	// TODO isdir
	t := time.Now()
	logFname := fmt.Sprintf("ci-builder-%d%02d%02d-%02d%02d%02d.log",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	daemonLog, err := os.Create(path.Join(config.LogPath, logFname))

	if err != nil {
		fmt.Printf("Could not open log file!! %s\n", err.Error())
		panic(err)
	}

	mw := io.MultiWriter(daemonLog, os.Stdout)
	Log = log.New(mw, "", log.Ldate|log.Ltime|log.Lshortfile)

	fmt.Printf("%+v\n", config)
	go processNotifications()
	hookEndpoint()
	/*
			repo := Repo{
				Name:           "TestRepo",
				GitUrl:         "ssh://git@gogs.davidventura.com.ar:2222/tati/critter-crossing.git",
				Bucket:         "testbucket",
				TelegramChatId: 0,
			}
			hookData := HookData{
				GitUrl: "ssh://git@gogs.davidventura.com.ar:2222/tati/critter-crossing.git",
				Ref:    "master",
			}
		buildRepo(repo, hookData)
	*/
}
