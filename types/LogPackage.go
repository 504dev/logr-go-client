package types

import (
	"encoding/base64"
	"encoding/json"
	"github.com/504dev/logr-go-client/helpers"
)

type LogPackage struct {
	DashId      int                    `json:"dash_id,omitempty"`
	PublicKey   string                 `json:"public_key"`
	CipherLog   string                 `json:"cipher_log,omitempty"`
	CipherCount string                 `json:"cipher_count,omitempty"`
	PlainLog    string                 `json:"_log,omitempty"`
	*Log        `json:"log,omitempty"` // deprecated field, do not support long messages
	*Count      `json:"count,omitempty"`
	ChunkUid    string `json:"uid,omitempty"`
	ChunkI      int    `json:"i,omitempty"`
	ChunkN      int    `json:"n,omitempty"`
}

func (lp *LogPackage) SerializeLog() error {
	msg, err := json.Marshal(lp.Log)
	if err != nil {
		return err
	}
	lp.PlainLog = base64.StdEncoding.EncodeToString(msg)
	lp.Log = nil
	return nil
}

func (lp *LogPackage) DeserializeLog() error {
	log := Log{}
	decoded, _ := base64.StdEncoding.DecodeString(lp.PlainLog)
	err := json.Unmarshal(decoded, &log)
	if err != nil {
		return err
	}
	lp.Log = &log
	return nil
}

func (lp *LogPackage) EncryptLog(priv string) error {
	cipherLog, err := lp.Log.Encrypt(priv)
	if err != nil {
		return err
	}
	lp.CipherLog = cipherLog
	lp.Log = nil
	return nil
}
func (lp *LogPackage) DecryptLog(priv string) error {
	log := Log{}
	err := log.Decrypt(lp.CipherLog, priv)
	if err != nil {
		return err
	}
	lp.Log = &log
	return nil
}

func (lp *LogPackage) EncryptCount(priv string) error {
	cipherText, err := lp.Count.Encrypt(priv)
	if err != nil {
		return err
	}
	lp.CipherCount = cipherText
	lp.Count = nil
	return nil
}
func (lp *LogPackage) DecryptCount(priv string) error {
	count := Count{}
	err := count.Decrypt(lp.CipherCount, priv)
	if err != nil {
		return err
	}
	lp.Count = &count
	return nil
}

func (lp *LogPackage) Chunkify(n int) ([][]byte, error) {
	msg, err := json.Marshal(lp)
	if err != nil {
		return nil, err
	}

	if len(msg) <= n {
		return [][]byte{msg}, err
	}

	var data string
	if lp.CipherLog != "" {
		data = lp.CipherLog
	} else {
		data = lp.PlainLog
	}
	headSize := len(msg) - len(data)
	chunkSize := n - headSize
	chunks := helpers.ChunkifyString(data, chunkSize)
	result := make([][]byte, len(chunks))
	uid := helpers.RandString(6)

	for i, chunk := range chunks {
		lpi := *lp
		lpi.ChunkUid = uid
		lpi.ChunkI = i
		lpi.ChunkN = len(result)
		if lp.CipherLog != "" {
			lpi.CipherLog = chunk
		} else {
			lpi.PlainLog = chunk
		}
		result[i], _ = json.Marshal(lpi)
	}

	return result, nil
}
