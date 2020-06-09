package logr_go_client

import (
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
)

func initiator() string {
	stack := string(debug.Stack())
	caller := strings.TrimSpace(strings.Split(stack, "\n")[12])
	splitted := regexp.MustCompile(`[\s\/]+`).Split(caller, 20)
	length := len(splitted)
	caller = strings.Join(splitted[length-3:length-1], "/")
	return caller
}

func readCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	stdout, _ := cmd.Output()

	return string(stdout)
}

func readTag() string {
	cmd := exec.Command("git", "tag", "-l", "--points-at", "HEAD")
	stdout, _ := cmd.Output()
	tmp := string(stdout)
	parts := strings.Split(tmp, "\n")

	if len(parts) > 1 {
		return parts[len(parts)-2]
	}
	return ""
}
