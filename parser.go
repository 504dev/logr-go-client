package go_kidlog

import (
	"github.com/504dev/kidlog/types"
)

type Parser struct {
	*Logger
	Handle func(log *Log)
}

func (p *Parser) Write(b []byte) (int, error) {
	log := p.blankLog()
	log.Level = LevelInfo
	log.Message = string(b)

	if p.Handle != nil {
		p.Handle(&Log{Log: log})
	}

	return p.writeLog(log)
}

type Log struct {
	*types.Log
}
