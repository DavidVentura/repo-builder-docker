package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var logTemplate = template.Must(template.New("logTemplate").Parse(`
	<html>
	<head><meta http-equiv='refresh' content='3'></head>
	<body><pre>{{.}}</pre></body>
	</html>`))

func hookEndpoint() {
	http.HandleFunc("/hook", hookHandler)
	http.HandleFunc("/logs/", logHandler)
	http.ListenAndServe(":8080", nil)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	logFile := strings.TrimPrefix(r.URL.Path, "/logs/")

	path := path.Join(config.LogPath, logFile)
	if _, err := os.Stat(path); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	content, err := ioutil.ReadAll(file)
	//contentStr := strings.ReplaceAll(string(content), "\n", "<br/>")
	contentStr := string(content)
	logTemplate.Execute(w, contentStr)

}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header["X-Gogs-Event"][0] != "create" {
		Log.Printf("Event was not 'create', refusing to work")
		Log.Printf("Headers: %+v", r.Header)
		return
	}

	var event GogsPushEvent
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &event)
	Log.Printf("Parsed body of request: %+v", event)
	if event.Ref_type != "tag" {
		Log.Printf("Received an event that is not from a tag. Aborting")
		return
	}
	for _, repo := range config.Repos {
		if repo.GitUrl == event.Repository.Ssh_url {
			Log.Printf("Repo %+v matches request.", repo)
			hookData := HookData{
				GitUrl: repo.GitUrl,
				Tag:    event.Ref,
			}
			go buildRepo(repo, hookData)
		}
	}
}

func buildRepo(repo Repo, hookData HookData) {
	repo.dstPath = path.Join(config.RepoCloneBase, repo.Name)
	logName := fmt.Sprintf("%s-%d.log", repo.Name, time.Now().Nanosecond())
	logPath := path.Join(config.LogPath, logName)
	buildLog, err := os.Create(logPath)

	Log.Printf("Cloning repo %+v, you can find the log at %s", repo, logPath)

	if err != nil {
		Log.Printf("[E] Could not open log file for %+v: %s", repo, err.Error())
		return
	}

	defer buildLog.Close()
	defer buildLog.Sync()

	err = cloneRepo(repo, hookData, buildLog)
	if err != nil {
		buildLog.Write([]byte("Failed to clone repo!\n"))
		buildLog.Write([]byte(err.Error()))
		return
	}
	logUrl := fmt.Sprintf("http://ci.labs/logs/%s", logName)
	notifications <- fmt.Sprintf("Starting build for %s, you can find the logs at %s", repo.Name, logUrl)
	err = dockerBuild(repo, hookData, buildLog)
	if err != nil {
		Log.Printf("Failed building repo %+v", repo)
		notifications <- "Build failed!"
		buildLog.Write([]byte(fmt.Sprintf("Failed to build repo!\n%s\n", err.Error())))
		return
	}
	notifications <- fmt.Sprintf("Build of %s@%s succeeded!", repo.Name, hookData.Tag)
	Log.Printf("Finished building repo %+v", repo)
}
