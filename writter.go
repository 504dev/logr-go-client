package logr_go_client

import (
	"github.com/504dev/logr-go-client/types"
)

type Writer struct {
	*Logger
	Transform func(log *Log)
}

func (w *Writer) Write(b []byte) (int, error) {
	log := w.blankLog()
	log.Level = types.LevelInfo
	log.Message = string(b)

	if w.Transform != nil {
		w.Transform(&Log{Log: log})
	}

	return w.PushLog(log)
}

type Log struct {
	*types.Log
}
