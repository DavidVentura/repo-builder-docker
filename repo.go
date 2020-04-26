package main

import (
	"encoding/json"
	"os"
	"path"
)

func readRepoBuildConfig(repopath string) (*RepoBuildConfig, error) {
	file, err := os.Open(path.Join(repopath, "build.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var repoConfig *RepoBuildConfig
	err = decoder.Decode(&repoConfig)
	if err != nil {
		return nil, err
	}

	return repoConfig, nil
}
