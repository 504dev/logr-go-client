package go_kidlog

import (
	"github.com/504dev/kidlog/types"
)

type Writter struct {
	*Logger
	Transform func(log *Log)
}

func (p *Writter) Write(b []byte) (int, error) {
	log := p.blankLog()
	log.Level = LevelInfo
	log.Message = string(b)

	if p.Transform != nil {
		p.Transform(&Log{Log: log})
	}

	return p.writeLog(log)
}

type Log struct {
	*types.Log
}
