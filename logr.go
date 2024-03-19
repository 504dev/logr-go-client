package logr_go_client

import (
	"fmt"
	"github.com/504dev/logr-go-client/types"
	"github.com/504dev/logr-go-client/utils"
	"github.com/fatih/color"
	"strconv"
	"strings"
	"time"
)

type levels struct {
	Emerg  types.Level
	Alert  types.Level
	Crit   types.Level
	Error  types.Level
	Warn   types.Level
	Notice types.Level
	Info   types.Level
	Debug  types.Level
}

var Levels = levels{
	Emerg:  types.LevelEmerg,
	Alert:  types.LevelAlert,
	Crit:   types.LevelCrit,
	Error:  types.LevelError,
	Warn:   types.LevelWarn,
	Notice: types.LevelNotice,
	Info:   types.LevelInfo,
	Debug:  types.LevelDebug,
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
	Levels levels
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

func (lg *Logger) prefix(level types.Level) string {
	dt := time.Now().Format(time.RFC3339)
	var colored string
	switch level {
	case types.LevelEmerg:
		fallthrough
	case types.LevelAlert:
		fallthrough
	case types.LevelCrit:
		colored = colorCrit(level)
	case types.LevelError:
		colored = colorError(level)
	case types.LevelWarn:
		colored = colorWarn(level)
	case types.LevelNotice:
		colored = colorNotice(level)
	case types.LevelInfo:
		colored = colorInfo(level)
	case types.LevelDebug:
		colored = colorDebug(level)
	default:
		colored = string(level)
	}
	res := lg.Prefix
	res = strings.Replace(res, "{time}", dt, -1)
	res = strings.Replace(res, "{level}", colored, -1)
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

func (lg *Logger) InfoErr(err error, v ...interface{}) {
	if err == nil {
		lg.Log(types.LevelInfo, v...)
	} else {
		lg.Error(types.LevelInfo, v...)
	}
}

func (lg *Logger) Emerg(v ...interface{}) {
	lg.Log(types.LevelEmerg, v...)
}

func (lg *Logger) Alert(v ...interface{}) {
	lg.Log(types.LevelAlert, v...)
}

func (lg *Logger) Crit(v ...interface{}) {
	lg.Log(types.LevelCrit, v...)
}

func (lg *Logger) Error(v ...interface{}) {
	lg.Log(types.LevelError, v...)
}

func (lg *Logger) Warn(v ...interface{}) {
	lg.Log(types.LevelWarn, v...)
}

func (lg *Logger) Notice(v ...interface{}) {
	lg.Log(types.LevelNotice, v...)
}

func (lg *Logger) Info(v ...interface{}) {
	lg.Log(types.LevelInfo, v...)
}

func (lg *Logger) Debug(v ...interface{}) {
	lg.Log(types.LevelDebug, v...)
}

func (lg *Logger) Log(level types.Level, v ...interface{}) {
	if lg.Level != "" && level.Weight() < types.Level(lg.Level).Weight() {
		return
	}
	prefix := lg.prefix(level)
	body := lg.body(format(v...))
	if lg.Console {
		fmt.Fprintln(level.Std(), prefix+body)
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

func (lg *Logger) writeLevel(level types.Level, msg string) (int, error) {
	log := lg.blankLog()
	log.Level = string(level)
	log.Message = msg

	return lg.PushLog(log)
}
