package types

import (
	"github.com/504dev/logr-go-client/cipher"
)

type Log struct {
	DashId    int    `db:"dash_id"   json:"dash_id"`
	Timestamp int64  `db:"timestamp" json:"timestamp,string"`
	Hostname  string `db:"hostname"  json:"hostname"`
	Logname   string `db:"logname"   json:"logname"`
	Level     string `db:"level"     json:"level"`
	Message   string `db:"message"   json:"message"`
	Pid       int    `db:"pid"       json:"pid"`
	Version   string `db:"version"   json:"version,omitempty"`
	Initiator string `db:"initiator" json:"initiator,omitempty"`
}

type Logs []*Log

func (log *Log) Decrypt(cipherBytes []byte, priv string) error {
	return cipher.DecodeAesJson(cipherBytes, priv, log)
}

func (log *Log) Encrypt(priv string) ([]byte, error) {
	return cipher.EncryptAesJson(log, priv)
}
