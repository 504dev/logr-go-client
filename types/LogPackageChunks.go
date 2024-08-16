package types

import (
	"bytes"
	gojson "github.com/goccy/go-json"
)

type LogPackageChunks []*LogPackage

func (chunks LogPackageChunks) isComplete() bool {
	if len(chunks) == 0 {
		return false
	}
	for _, lp := range chunks {
		if lp == nil {
			return false
		}
	}
	return true
}

func (chunks LogPackageChunks) Joined() (complete bool, joined *LogPackage) {
	if !chunks.isComplete() {
		return false, nil
	}

	ciphered := chunks[0].CipherLog != nil
	var buffer bytes.Buffer

	for _, lp := range chunks {
		if ciphered {
			buffer.Write(lp.CipherLog)
		} else {
			buffer.Write(lp.PlainLog)
		}
	}

	clone := *chunks[0]
	if ciphered {
		clone.CipherLog = buffer.Bytes()
	} else {
		clone.PlainLog = buffer.Bytes()
	}

	return true, &clone
}

func (chunks LogPackageChunks) Marshal() ([][]byte, error) {
	result := make([][]byte, len(chunks))
	for i, lp := range chunks {
		msg, err := gojson.Marshal(lp)
		if err != nil {
			return nil, err
		}
		result[i] = msg
	}
	return result, nil
}
