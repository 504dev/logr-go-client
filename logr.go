package logr_go_client

import (
	"fmt"
	"github.com/504dev/logr-go-client/types"
	"github.com/504dev/logr-go-client/utils"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	LevelEmerg  = "emerg"
	LevelAlert  = "alert"
	LevelCrit   = "crit"
	LevelError  = "error"
	LevelWarn   = "warn"
	LevelNotice = "notice"
	LevelInfo   = "info"
	LevelDebug  = "debug"
)

var weights = map[string]int{
	LevelEmerg:  7,
	LevelAlert:  6,
	LevelCrit:   5,
	LevelError:  4,
	LevelWarn:   3,
	LevelNotice: 2,
	LevelInfo:   1,
	LevelDebug:  0,
}

var std = map[string]*os.File{
	LevelEmerg:  os.Stderr,
	LevelAlert:  os.Stderr,
	LevelCrit:   os.Stderr,
	LevelError:  os.Stderr,
	LevelWarn:   os.Stderr,
	LevelNotice: os.Stdout,
	LevelInfo:   os.Stdout,
	LevelDebug:  os.Stdout,
}

const MAX_MESSAGE_SIZE = 9000

type Logger struct {
	*Config
	Transport
	Logname string
	Body    string
	Prefix  string
	Level   string
	Console bool
	*Counter
}

func (lg *Logger) Close() error {
	err := lg.Transport.Close()
	if err != nil {
		return err
	}
	return lg.Counter.Transport.Close()
}

func (lg *Logger) Of(logname string) *Logger {
	tmp := *lg
	tmp.Logname = logname
	return &tmp
}

func (lg *Logger) DefaultWriter() *Writer {
	return &Writer{
		Logger: lg,
	}
}

func (lg *Logger) CustomWriter(f func(log *Log)) *Writer {
	return &Writer{
		Logger:    lg,
		Transform: f,
	}
}

var colorCrit = color.New(color.FgRed).SprintFunc()
var colorError = color.New(color.FgHiRed).SprintFunc()
var colorWarn = color.New(color.FgYellow).SprintFunc()
var colorNotice = color.New(color.FgHiGreen).SprintFunc()
var colorInfo = color.New(color.FgGreen).SprintFunc()
var colorDebug = color.New(color.FgBlue).SprintFunc()

func (lg *Logger) prefix(level string) string {
	dt := time.Now().Format(time.RFC3339)
	flevel := level
	switch level {
	case LevelEmerg:
		fallthrough
	case LevelAlert:
		fallthrough
	case LevelCrit:
		flevel = colorCrit(level)
	case LevelError:
		flevel = colorError(level)
	case LevelWarn:
		flevel = colorWarn(level)
	case LevelNotice:
		flevel = colorNotice(level)
	case LevelInfo:
		flevel = colorInfo(level)
	case LevelDebug:
		flevel = colorDebug(level)
	}
	res := lg.Prefix
	res = strings.Replace(res, "{time}", dt, -1)
	res = strings.Replace(res, "{level}", flevel, -1)
	return res
}

func (lg *Logger) body(msg string) string {
	res := lg.Body
	initiator, caller := utils.Initiator()
	res = strings.Replace(res, "{logname}", lg.Logname, -1)
	res = strings.Replace(res, "{version}", lg.GetVersion(), -1)
	res = strings.Replace(res, "{pid}", strconv.Itoa(lg.GetPid()), -1)
	res = strings.Replace(res, "{initiator}", initiator, -1)
	res = strings.Replace(res, "{caller}", caller, -1)
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

func (lg *Logger) Emerg(v ...interface{}) {
	lg.Log(LevelEmerg, v...)
}

func (lg *Logger) Alert(v ...interface{}) {
	lg.Log(LevelAlert, v...)
}

func (lg *Logger) Crit(v ...interface{}) {
	lg.Log(LevelCrit, v...)
}

func (lg *Logger) Error(v ...interface{}) {
	lg.Log(LevelError, v...)
}

func (lg *Logger) Warn(v ...interface{}) {
	lg.Log(LevelWarn, v...)
}

func (lg *Logger) Notice(v ...interface{}) {
	lg.Log(LevelNotice, v...)
}

func (lg *Logger) Info(v ...interface{}) {
	lg.Log(LevelInfo, v...)
}

func (lg *Logger) Debug(v ...interface{}) {
	lg.Log(LevelDebug, v...)
}

func (lg *Logger) Log(level string, v ...interface{}) {
	if lg.Level != "" && weights[level] < weights[lg.Level] {
		return
	}
	prefix := lg.prefix(level)
	body := lg.body(format(v...))
	if lg.Console {
		fmt.Fprintln(std[level], prefix+body)
	}
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

	return lg.PushLog(log)
}
