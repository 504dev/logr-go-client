package go_kidlog

import (
	"encoding/json"
	"fmt"
	"github.com/504dev/go-kidlog/types"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

var commit = readCommit()
var pid = os.Getpid()

func caller() string {
	stack := string(debug.Stack())
	caller := strings.TrimSpace(strings.Split(stack, "\n")[10])
	splitted := regexp.MustCompile(`[\s\/]+`).Split(caller, 20)
	length := len(splitted)
	caller = strings.Join(splitted[length-3:length-1], "/")
	return caller
}

func prefix(level string) string {
	dt := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("[KID] %v %v [%v, pid=%v, %v]", dt, level, commit[:6], pid, caller())
}

func readCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err.Error)
		return ""
	}

	return string(stdout)
}

type Config struct {
	Udp        string
	PublicKey  string
	PrivateKey string
	Hostname   string
}

func (c *Config) Create(logname string) *Logger {
	return &Logger{
		Config:  c,
		Logname: logname,
	}
}

type Logger struct {
	*Config
	Logname string
}

func (lg *Logger) Info(v ...interface{}) {
	args := []interface{}{prefix("info")}
	args = append(args, v...)
	fmt.Println(args...)
	fmt.Fprintln(lg, args...)
}

func (lg *Logger) Error(v ...interface{}) {
	args := []interface{}{prefix("info")}
	args = append(args, v...)
	log.Println(args...)
}

func (lg Logger) Write(b []byte) (int, error) {
	conn, err := net.Dial("udp", "127.0.0.1:7776")
	if err != nil {
		return 0, err
	}
	logitem := types.Log{
		DashId:    0,
		Timestamp: time.Now().UnixNano(),
		Hostname:  lg.Config.Hostname,
		Logname:   lg.Logname,
		Level:     "info",
		Message:   string(b),
	}
	cipherText, err := logitem.Encrypt(lg.PrivateKey)
	if err != nil {
		return 0, err
	}
	lp := types.LogPackage{
		PublicKey:  lg.PublicKey,
		CipherText: cipherText,
	}
	msg, err := json.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = conn.Write(msg)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return len(b), nil
}
