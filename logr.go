package logr_go_client

import (
	"encoding/json"
	"fmt"
	"github.com/504dev/logr/types"
	"github.com/fatih/color"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

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

func (c *Config) NewLogger(logname string) (*Logger, error) {
	conn, err := net.Dial("udp", c.Udp)
	res := &Logger{
		Config:  c,
		Logname: logname,
		Prefix:  "{time} {level} ",
		Body:    "[{version}, pid={pid}, {initiator}] {message}",
		Conn:    conn,
		Counter: &Counter{
			Config:  c,
			Conn:    conn,
			Logname: logname,
			Tmp:     make(map[string]*types.Count),
		},
	}
	res.Counter.run(10 * time.Second)
	return res, err
}

type Logger struct {
	*Config
	net.Conn
	Logname string
	Body    string
	Prefix  string
	*Counter
}

func (lg *Logger) DefaultWritter() *Writter {
	return &Writter{
		Logger: lg,
	}
}

func (lg *Logger) CustomWritter(f func(log *Log)) *Writter {
	return &Writter{
		Logger:    lg,
		Transform: f,
	}
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
	res = strings.Replace(res, "{version}", lg.GetVersion(), -1)
	res = strings.Replace(res, "{pid}", strconv.Itoa(lg.GetPid()), -1)
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

func (lg *Logger) Log(level string, v ...interface{}) {
	prefix := lg.prefix(level)
	body := lg.body(format(v...))
	fmt.Fprintln(std[level], prefix+body)
	lg.writeLevel(level, body)
}

func (lg *Logger) blankLog() *types.Log {
	return &types.Log{
		DashId:    lg.Config.DashId,
		Timestamp: time.Now().UnixNano(),
		Hostname:  lg.GetHostname(),
		Logname:   lg.Logname,
		Pid:       lg.GetPid(),
		Version:   lg.GetVersion(),
	}
}

func (lg *Logger) writeLevel(level string, msg string) (int, error) {
	log := lg.blankLog()
	log.Level = level
	log.Message = msg

	return lg.writeLog(log)
}

func (lg *Logger) writeLog(log *types.Log) (int, error) {
	if lg.Conn == nil {
		return 0, nil
	}
	cipherLog, err := log.Encrypt(lg.PrivateKey)
	if err != nil {
		return 0, err
	}
	lp := types.LogPackage{
		DashId:    lg.DashId,
		PublicKey: lg.PublicKey,
		CipherLog: cipherLog,
	}
	msg, err := json.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = lg.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}
