package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

const directoryName = "gh-mirror"
const configFileName = "gh-mirror.json"

type Config struct {
	SleepDuration int         `json:"sleepDuration"`
	Repos         []RepoEntry `json:"repos"`
}

func getConfigPath() string {
	configPath := filepath.Join(getRootPath(), configFileName)
	return configPath
}

func createConfig() {
	var config Config
	config.SleepDuration = 30
	config.Repos = []RepoEntry{}
	writeConfig(config)
	fmt.Printf("Created config file: %v\n", getConfigPath())
}

func readConfig() Config {
	var config Config
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("could not parse config: %v", err)
	}
	return config
}

func writeConfig(config Config) {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		log.Fatalf("could not serialize config: %v", err)
	}
	err = os.WriteFile(configPath, data, os.ModePerm)
	if err != nil {
		log.Fatalf("could not write config: %v", err)
	}
}

func addRepoToConfig(repo RepoEntry) {
	fmt.Println("adding repo", repo.Display())
	config := readConfig()
	newRepo, err := repo.Resolve()
	if err != nil {
		log.Fatalf("invalid repo: %v", err)
	}
	for _, existing := range config.Repos {
		cfgRepo, err := existing.Resolve()
		if err != nil {
			log.Fatalf("invalid repo in config %q: %v", existing.Display(), err)
		}
		if cfgRepo.LocalPath == newRepo.LocalPath && cfgRepo.RemoteKey == newRepo.RemoteKey {
			// prevent duplicates
			return
		}
		if cfgRepo.LocalPath == newRepo.LocalPath {
			log.Fatalf("local path already exists for %v; set an explicit path to keep both: %v", newRepo.DisplayPath(), existing.Display())
		}
	}
	config.Repos = append(config.Repos, repo)
	sort.Slice(config.Repos, func(i, j int) bool {
		return config.Repos[i].SortKey() < config.Repos[j].SortKey()
	})
	writeConfig(config)
}

func getRootPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Fatal("$HOME is not set")
	}
	rootPath := filepath.Join(homeDir, directoryName)
	return rootPath
}

func createRoot() {
	rootPath := getRootPath()
	err := os.Mkdir(rootPath, os.ModePerm)
	if err != nil {
		log.Fatalf("could not create directory %v: %v", rootPath, err)
	}
	fmt.Printf("Created directory: %v\n", rootPath)
}
