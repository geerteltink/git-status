package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func main() {
	fmt.Println("Starting git status...")

	// If .git dir exists we're in a single repo
	if checkDir(".git") {
		// Display git status
		displayGitStatus(".")
		return
	}

	// Scan all sub directories
	fmt.Print("Scanning sub directories of . \n\n")
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		dir := file.Name()
		if checkDir(dir) {
			// Display git status
			os.Chdir(dir)
			displayGitStatus(dir)
			os.Chdir("..")
		}
	}
}

// Check dir if its a git repo
func checkDir(dir string) bool {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return false
	}

	if !fileInfo.IsDir() {
		return false
	}

	return true
}

func displayGitStatus(project string) {
	if !checkDir(".git") {
		// Not a git repository
		return
	}

	// project
	// branch := ""
	// ahead := 0
	// behind := 0
	// changedFiles := 0

	branch := getBranch()
	changes := getChanges()
	changedFiles := getChangedFiles()

	error := color.New(color.FgHiRed).SprintFunc()
	//errorLabel := color.New(color.BgRed, color.FgWhite).SprintFunc()
	//notice := color.New(color.FgBlue).SprintFunc()
	success := color.New(color.FgHiGreen).SprintFunc()

	if changes != "" || len(changedFiles) > 0 {
		project = error(project)
	} else {
		project = success(project)
	}

	if changes != "" {
		changes = strings.Replace(changes, "ahead ", "↑", -1)
		changes = strings.Replace(changes, "behind ", "↓", -1)
	}

	changedFilesStatus := ""
	if len(changedFiles) > 0 {
		changedFilesStatus = fmt.Sprintf("[+%d]", len(changedFiles))
	}

	fmt.Printf("%s/%s %s%s\n", project, branch, error(changes), error(changedFilesStatus))
}

func getBranch() string {
	// Get branch name
	cmdName := "git"
	cmdArgs := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	branch, err := exec.Command(cmdName, cmdArgs...).Output()
	if err == nil {
		return strings.TrimSpace(string(branch))
	}

	// Might be a new repo, fallback to status
	cmdName = "git"
	cmdArgs = []string{"status", "-bs"}
	branch, err = exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return "------"
	}

	// Strip spaces and hastags and return the string
	return strings.Trim(strings.TrimSpace(string(branch)), "# ")
}

func getChanges() string {
	cmdName := "git"
	cmdArgs := []string{"for-each-ref", "--format=%(push:track)", "refs/heads"}

	changes, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(changes))
}

func getChangedFiles() []string {
	changedFiles := []string{}
	cmdName := "git"
	cmdArgs := []string{"status", "-s"}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return changedFiles
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			//fmt.Println(scanner.Text())
			changedFiles = append(changedFiles, scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return changedFiles
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return changedFiles
	}

	return changedFiles
}
