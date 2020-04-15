package go_kidlog

import (
	"github.com/504dev/kidlog/types"
)

type Writter struct {
	*Logger
	Transform func(log *Log)
}

func (w *Writter) Write(b []byte) (int, error) {
	log := w.blankLog()
	log.Level = LevelInfo
	log.Message = string(b)

	if w.Transform != nil {
		w.Transform(&Log{Log: log})
	}

	return w.writeLog(log)
}

type Log struct {
	*types.Log
}
