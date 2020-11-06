package logr_go_client

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
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

func HtopTime() float64 {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("ps -eo time,pid | grep %v", os.Getpid()))
	stdout, err := cmd.Output()
	if err != nil {
		return 0
	}
	text := strings.TrimSpace(string(stdout))
	split := regexp.MustCompile("[\\s:]").Split(text, -1)
	if len(split) < 4 {
		split = append([]string{"0"}, split...)
	}
	hour, _ := strconv.ParseFloat(split[0], 8)
	min, _ := strconv.ParseFloat(split[1], 8)
	sec, _ := strconv.ParseFloat(split[2], 8)
	res := hour*60*60 + min*60 + sec
	return res
}
