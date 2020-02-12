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
		Body:    "[{tag|commit}, pid={pid}, {initiator}] {message}",
		Prefix:  "{time} {level} ",
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

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

var std = map[string]*os.File{
	LevelDebug: os.Stdout,
	LevelInfo:  os.Stdout,
	LevelWarn:  os.Stderr,
	LevelError: os.Stderr,
}

func (lg *Logger) prefix(level string) string {
	dt := time.Now().Format(time.RFC3339)
	flevel := level
	switch level {
	case LevelDebug:
		flevel = color.New(color.FgBlue).SprintFunc()(level)
	case LevelInfo:
		flevel = color.New(color.FgGreen).SprintFunc()(level)
	case LevelWarn:
		flevel = color.New(color.FgYellow).SprintFunc()(level)
	case LevelError:
		flevel = color.New(color.FgRed).SprintFunc()(level)
	}
	res := lg.Prefix
	res = strings.Replace(res, "{time}", dt, -1)
	res = strings.Replace(res, "{level}", flevel, -1)
	return res
}

func (lg *Logger) body(msg string) string {
	res := lg.Body
	commit := commit[:6]
	tagOrCommit := tag
	if tagOrCommit == "" {
		tagOrCommit = commit
	}
	res = strings.Replace(res, "{tag}", tag, -1)
	res = strings.Replace(res, "{commit}", commit, -1)
	res = strings.Replace(res, "{tag|commit}", tagOrCommit, -1)
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

func (lg *Logger) Log(level string, v ...interface{}) {
	prefix := lg.prefix(level)
	body := lg.body(format(v...))
	fmt.Fprintln(std[level], prefix+body)
	lg.writeLevel(level, []byte(body))
}

func (lg *Logger) Debug(v ...interface{}) {
	lg.Log(LevelDebug, v...)
}

func (lg *Logger) Info(v ...interface{}) {
	lg.Log(LevelInfo, v...)
}

func (lg *Logger) Warn(v ...interface{}) {
	lg.Log(LevelWarn, v...)
}

func (lg *Logger) Error(v ...interface{}) {
	lg.Log(LevelError, v...)
}

func (lg *Logger) SendLog(level string, v ...interface{}) (int, error) {
	body := lg.body(format(v...))
	return lg.writeLevel(level, []byte(body))
}

func (lg *Logger) Write(b []byte) (int, error) {
	return lg.writeLevel("info", b)
}

func (lg *Logger) writeLevel(level string, b []byte) (int, error) {

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
