package types

import "os"

const (
	LevelEmerg  Level = "emerg"
	LevelAlert        = "alert"
	LevelCrit         = "crit"
	LevelError        = "error"
	LevelWarn         = "warn"
	LevelNotice       = "notice"
	LevelInfo         = "info"
	LevelDebug        = "debug"
)

type Level string

func (lvl Level) Weight() int {
	return map[Level]int{
		LevelEmerg:  7,
		LevelAlert:  6,
		LevelCrit:   5,
		LevelError:  4,
		LevelWarn:   3,
		LevelNotice: 2,
		LevelInfo:   1,
		LevelDebug:  0,
	}[lvl]
}

func (lvl Level) Std() *os.File {
	std := map[Level]*os.File{
		LevelEmerg:  os.Stderr,
		LevelAlert:  os.Stderr,
		LevelCrit:   os.Stderr,
		LevelError:  os.Stderr,
		LevelWarn:   os.Stderr,
		LevelNotice: os.Stdout,
		LevelInfo:   os.Stdout,
		LevelDebug:  os.Stdout,
	}[lvl]
	if std == nil {
		std = os.Stdout
	}
	return std
}
