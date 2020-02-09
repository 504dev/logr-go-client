package go_kidlog

import (
	"encoding/json"
	"fmt"
	"github.com/504dev/kidlog/types"
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
	return fmt.Sprintf("[KID] %v %v [%v, pid=%v, %v] ", dt, level, commit[:6], pid, caller())
}

func readCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
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

func (c *Config) Create(logname string) (*Logger, error) {
	conn, err := net.Dial("udp", "127.0.0.1:7776")
	if err != nil {
		return nil, err
	}
	res := &Logger{
		Config:  c,
		Logname: logname,
		Conn:    conn,
	}
	return res, nil
}

type Logger struct {
	*Config
	Logname string
	net.Conn
}

func (lg *Logger) Format(level string, vals ...interface{}) string {
	pfx := prefix(level)
	switch v := vals[0].(type) {
	case string:
		return fmt.Sprintf(pfx+v, vals[1:]...)
	default:
		args := []interface{}{pfx}
		args = append(args, vals...)
		return fmt.Sprint(args...)
	}
}

func (lg *Logger) Info(v ...interface{}) {
	level := "info"
	formatted := lg.Format(level, v...)
	fmt.Fprintln(os.Stdout, formatted)
	lg.WriteLevel(level, []byte(formatted))
}

func (lg *Logger) Error(v ...interface{}) {
	level := "error"
	formatted := lg.Format(level, v...)
	fmt.Fprintln(os.Stderr, formatted)
	lg.WriteLevel(level, []byte(formatted))
}

func (lg *Logger) Warn(v ...interface{}) {
	level := "warn"
	formatted := lg.Format(level, v...)
	fmt.Fprintln(os.Stderr, formatted)
	lg.WriteLevel(level, []byte(formatted))
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
