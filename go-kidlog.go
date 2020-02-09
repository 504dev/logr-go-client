package go_kidlog

import (
	"encoding/json"
	"fmt"
	"github.com/504dev/kidlog/types"
	"github.com/fatih/color"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var commit = readCommit()
var tag = readTag()
var pid = os.Getpid()

func initiator() string {
	stack := string(debug.Stack())
	caller := strings.TrimSpace(strings.Split(stack, "\n")[10])
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

	return string(stdout)
}

type Config struct {
	Udp        string
	PublicKey  string
	PrivateKey string
	Hostname   string
}

func (c *Config) Create(logname string) (*Logger, error) {
	conn, err := net.Dial("udp", "127.0.0.1:7776")
	if err != nil {
		return nil, err
	}
	res := &Logger{
		Config:  c,
		Logname: logname,
		Body:    "[{tag}, {commit}, pid={pid}, {initiator}] {message}",
		Prefix:  "{time} {level}",
		Conn:    conn,
	}
	return res, nil
}

type Logger struct {
	*Config
	Logname string
	Body    string
	Prefix  string
	net.Conn
}

func (lg *Logger) prefix(level string) string {
	dt := time.Now().Format(time.RFC3339)
	flevel := level
	switch level {
	case "info":
		flevel = color.New(color.FgGreen).SprintFunc()(level)
	case "warn":
		flevel = color.New(color.FgYellow).SprintFunc()(level)
	case "error":
		flevel = color.New(color.FgRed).SprintFunc()(level)
	}
	res := lg.Prefix
	res = strings.Replace(res, "{time}", dt, -1)
	res = strings.Replace(res, "{level}", flevel, -1)
	return res
}

func (lg *Logger) body(msg string) string {
	res := lg.Body
	res = strings.Replace(res, "{tag}", tag, -1)
	res = strings.Replace(res, "{commit}", commit[:6], -1)
	res = strings.Replace(res, "{pid}", strconv.Itoa(pid), -1)
	res = strings.Replace(res, "{initiator}", initiator(), -1)
	res = strings.Replace(res, "{message}", msg, -1)
	return res
}

func format(vals ...interface{}) string {
	switch v := vals[0].(type) {
	case string:
		return fmt.Sprintf(v, vals[1:]...)
	default:
		return fmt.Sprint(vals...)
	}
}

func (lg *Logger) Info(v ...interface{}) {
	level := "info"
	msg := format(v...)
	fmt.Fprintln(os.Stdout, lg.prefix(level)+lg.body(msg))
	lg.WriteLevel(level, []byte(lg.body(msg)))
}

func (lg *Logger) Error(v ...interface{}) {
	level := "error"
	msg := format(v...)
	fmt.Fprintln(os.Stderr, lg.prefix(level)+lg.body(msg))
	lg.WriteLevel(level, []byte(lg.body(msg)))
}

func (lg *Logger) Warn(v ...interface{}) {
	level := "warn"
	msg := format(v...)
	fmt.Fprintln(os.Stderr, lg.prefix(level)+lg.body(msg))
	lg.WriteLevel(level, []byte(lg.body(msg)))
}

func (lg *Logger) Write(b []byte) (int, error) {
	return lg.WriteLevel("info", b)
}

func (lg *Logger) WriteLevel(level string, b []byte) (int, error) {

	logitem := types.Log{
		DashId:    0,
		Timestamp: time.Now().UnixNano(),
		Hostname:  lg.Config.Hostname,
		Logname:   lg.Logname,
		Level:     level,
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
	_, err = lg.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
