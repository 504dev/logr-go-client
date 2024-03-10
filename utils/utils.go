package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
)

func ReadGitCommit() string {
	commit := ReadGitCommitDir("")
	if commit == "" {
		execFile, _ := os.Executable()
		commit = ReadGitCommitDir(filepath.Dir(execFile))
	}
	return commit
}

func ReadGitCommitDir(dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		cmd := exec.Command("cat", ".git/HEAD")
		cmd.Dir = dir
		stdout, _ = cmd.Output()
	}
	commit := strings.TrimSuffix(string(stdout), "\n")

	return commit
}

func ReadGitTag() string {
	return ReadGitTagDir("")
}

func ReadGitTagDir(dir string) string {
	cmd := exec.Command("git", "tag", "-l", "--points-at", "HEAD")
	cmd.Dir = dir
	stdout, _ := cmd.Output()
	tmp := string(stdout)
	parts := strings.Split(tmp, "\n")
	if len(parts) > 1 {
		return parts[len(parts)-2]
	}
	return ""
}

func Initiator() (string, string) {
	stack := string(debug.Stack())
	stackSplitted := strings.Split(stack, "\n")
	caller := strings.TrimSpace(stackSplitted[11])
	initiator := strings.TrimSpace(stackSplitted[12])

	start, end := strings.LastIndex(caller, "/"), strings.LastIndex(caller, "(")
	callerSplitted := strings.Split(caller[start+1:end], ".")
	if length := len(callerSplitted); length > 2 {
		callerSplitted = callerSplitted[length-2 : length]
	}
	caller = strings.Join(callerSplitted, ".")

	initiatorSplitted := regexp.MustCompile(`[\s/]+`).Split(initiator, 20)
	length := len(initiatorSplitted)
	initiator = strings.Join(initiatorSplitted[length-3:length-1], "/")

	return initiator, caller
}
