package types

import (
	"github.com/504dev/go-kidlog/cipher"
)

type Log struct {
	DashId    int    `db:"dash_id"   json:"dash_id"`
	Timestamp int64  `db:"timestamp" json:"timestamp"`
	Hostname  string `db:"hostname"  json:"hostname"`
	Logname   string `db:"logname"   json:"logname"`
	Level     string `db:"level"     json:"level"`
	Message   string `db:"message"   json:"message"`
}

type Metr struct {
	DashId    int                `db:"dash_id"   json:"dash_id"`
	Timestamp int64              `db:"timestamp" json:"timestamp"`
	Hostname  string             `db:"hostname"  json:"hostname"`
	Type      string             `db:"type"      json:"type"`
	Name      string             `db:"name"      json:"name"`
	Scores    map[string]float64 `db:"scores"    json:"scores"`
}

type Logs []*Log

func (log *Log) Decrypt(cipherText string, priv string) error {
	return cipher.DecodeAesJson(cipherText, priv, log)
}

func (log *Log) Encrypt(priv string) (string, error) {
	return cipher.EncryptAesJson(log, priv)
}

type LogPackage struct {
	PublicKey  string `json:"publickey"`
	CipherText string `json:"ciphertext"`
	*Log       `json:"log"`
	*Metr      `json:"metr"`
}
