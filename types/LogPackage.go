package types

import (
	"encoding/json"
	"github.com/504dev/logr-go-client/helpers"
)

type LogPackage struct {
	DashId      int    `json:"dash_id,omitempty"`
	PublicKey   string `json:"public_key"`
	CipherLog   string `json:"cipher_log,omitempty"`
	CipherCount string `json:"cipher_count,omitempty"`
	*Log        `json:"log,omitempty"`
	*Count      `json:"count,omitempty"`
	ChunkUid    string `json:"uid,omitempty"`
	ChunkI      int    `json:"i,omitempty"`
	ChunkN      int    `json:"n,omitempty"`
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

	if len(msg) <= n || lp.CipherLog == "" {
		return [][]byte{msg}, err
	}

	data := lp.CipherLog
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
		lpi.CipherLog = chunk
		result[i], _ = json.Marshal(lpi)
	}

	return result, nil
}
