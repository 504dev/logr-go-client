package go_kidlog

import (
	"github.com/504dev/kidlog/types"
)

type Parser struct {
	*Logger
	Handle func(log *Log)
}

func (p *Parser) Write(b []byte) (int, error) {
	blank := p.blankLog()
	blank.Level = LevelInfo
	blank.Message = string(b)

	p.Handle(&Log{Log: blank})

	return p.writeLog(blank)
}

type Log struct {
	*types.Log
}
