package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	if r.Header["X-Gogs-Event"][0] != "create" && r.Header["X-Gogs-Event"][0] != "push" {
		Log.Printf("Event was not 'create' nor 'push', refusing to work")
		Log.Printf("Headers: %+v", r.Header)
		return
	}

	var event GogsPushEvent
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &event)
	Log.Printf("Parsed body of request: %+v", event)
	for _, repo := range config.Repos {
		if repo.GitUrl == event.Repository.Ssh_url {
			Log.Printf("Repo %+v matches request.", repo)
			hookData := HookData{
				GitUrl: repo.GitUrl,
				Ref:    strings.Replace(event.Ref, "refs/heads/", "", 1),
			}
			go buildRepo(repo, hookData)
		}
	}
}

func buildRepo(repo Repo, hookData HookData) {
	repo.dstPath = path.Join(config.RepoCloneBase, repo.Name)
	logName := fmt.Sprintf("%s-%d.log", repo.Name, time.Now().Nanosecond())
	logPath := path.Join(config.LogPath, logName)
	buildLogFile, err := os.Create(logPath)
	buildLog := io.MultiWriter(buildLogFile, os.Stdout)

	Log.Printf("Cloning repo %+v, you can find the log at %s", repo, logPath)

	if err != nil {
		Log.Printf("[E] Could not open log file for %+v: %s", repo, err.Error())
		return
	}

	defer buildLogFile.Close()
	defer buildLogFile.Sync()

	err = cloneRepo(repo, hookData, buildLog)
	if err != nil {
		buildLog.Write([]byte("Failed to clone repo!\n"))
		buildLog.Write([]byte(err.Error()))
		buildLog.Write([]byte("\n"))
		return
	}
	logUrl := fmt.Sprintf("http://ci.labs/logs/%s", url.PathEscape(logName))

	repoBuildConfig, err := readRepoBuildConfig(repo.dstPath)
	if err != nil {
		buildLog.Write([]byte("Failed to parse the build.json in the repo!\n"))
		buildLog.Write([]byte(err.Error()))
		return
	}
	prettyRef := ""
	if hookData.Ref != "master" {
		prettyRef = fmt.Sprintf("@%s", hookData.Ref)
	}

	notifications <- Notification{
		msg:     fmt.Sprintf("Starting build for %s%s, you can find the logs at %s", repo.Name, prettyRef, logUrl),
		chat_id: repo.TelegramChatId}

	for _, subproject := range repoBuildConfig.Subprojects {

		err = dockerBuild(repo, hookData, buildLog, subproject)
		if err != nil {
			Log.Printf("Failed building repo %+v", repo)
			notifications <- Notification{msg: "Build failed!", chat_id: repo.TelegramChatId}
			buildLog.Write([]byte(fmt.Sprintf("Failed to build %s.%s!\n%s\n", repo.Name, subproject.Name, err.Error())))
			return
		}

	}
	notifications <- Notification{
		msg:     fmt.Sprintf("Build of %s%s succeeded!", repo.Name, prettyRef),
		chat_id: repo.TelegramChatId}
	Log.Printf("Finished building repo %+v", repo)

}
