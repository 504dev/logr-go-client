package utils

import (
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
)

func ReadGitCommit() string {
	return ReadGitCommitDir("")
}

func ReadGitCommitDir(dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	stdout, _ := cmd.Output()
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

func Initiator() string {
	stack := string(debug.Stack())
	caller := strings.TrimSpace(strings.Split(stack, "\n")[12])
	splitted := regexp.MustCompile(`[\s\/]+`).Split(caller, 20)
	length := len(splitted)
	caller = strings.Join(splitted[length-3:length-1], "/")
	return caller
}
