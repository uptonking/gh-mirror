package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	flInit := flag.Bool("init", false, "init gh-mirror")
	flAdd := flag.String("add", "", "add repo to gh-mirror (username/repo)")
	flList := flag.Bool("list", false, "list repos in gh-mirror")
	flag.Parse()
	log.SetFlags(0)

	if *flInit {
		initCmd()
		return
	}

	if *flAdd != "" {
		addCmd(*flAdd)
		return
	}

	if *flList {
		listCmd()
		return
	}

	baseCmd()
}

func initCmd() {
	createRoot()
	createConfig()
}

func addCmd(repoArg string) {
	repo, err := parseRepoArg(repoArg)
	if err != nil {
		log.Printf("usage: gh-mirror -add username/repo")
		log.Printf("   or: gh-mirror -add git@example.com:org/repo.git")
		log.Printf("   or: gh-mirror -add https://example.com/org/repo")
		log.Fatalf("%v", err)
	}
	addRepoToConfig(repo)
}

func listCmd() {
	config := readConfig()
	for _, repo := range config.Repos {
		fmt.Println(repo.Display())
	}
}

func baseCmd() {
	config := readConfig()
	for i, repo := range config.Repos {
		resolved, err := repo.Resolve()
		if err != nil {
			log.Fatalf("repo is invalid: %v", err)
		}
		log.SetPrefix(fmt.Sprintf("[%s]: ", resolved.DisplayPath()))

		if i > 0 {
			log.Printf("sleeping for %d seconds...\n", config.SleepDuration)
			time.Sleep(time.Duration(config.SleepDuration) * time.Second)
		}

		if notExistRepo(resolved) {
			cloneRepo(resolved)
		} else {
			updateRepo(resolved)
		}
	}
}

func cloneRepo(repo ResolvedRepo) {
	changeWorkDirToRoot()
	cloneCmd := "git"
	cloneArgs := []string{"clone", "-v", repo.CloneURL, repo.LocalPath}
	runCommand(cloneCmd, cloneArgs)
}

func updateRepo(repo ResolvedRepo) {
	changeWorkDirToRepo(repo)
	updateCmd := "git"
	// updateArgs := []string{"remote", "-v", "update"}
	//--rebase[=(false|true|merges|interactive)]
	updateArgs := []string{"pull", "-r"}
	runCommand(updateCmd, updateArgs)
}

func changeWorkDirToRoot() {
	dir := getRootPath()
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("could not change work directory: %v", err)
	}
	log.Printf("change work directory to %v\n", dir)
}

func changeWorkDirToRepo(repo ResolvedRepo) {
	path := filepath.Join(getRootPath(), repo.LocalPath)
	err := os.Chdir(path)
	if err != nil {
		log.Fatalf("could not change work directory: %v", err)
	}
	log.Printf("change work directory to %v\n", path)
}

func runCommand(cmd string, args []string) {
	log.Printf("%s %s\n", cmd, strings.Join(args, " "))
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		log.Printf("%s\v", string(output))
		log.Fatalf("fatal: could not run command: %v", err)
	}
	log.Println(string(output))
}

func notExistRepo(repo ResolvedRepo) bool {
	path := filepath.Join(getRootPath(), repo.LocalPath)
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

var reValidRepoPath = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9\._-]+$`)

func isValidRepoPath(repo string) bool {
	return reValidRepoPath.MatchString(repo)
}
