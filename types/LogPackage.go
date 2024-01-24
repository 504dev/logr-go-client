package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/504dev/logr-go-client/cipher"
	"github.com/504dev/logr-go-client/helpers"
	"time"
)

type LogPackage struct {
	DashId      int                    `json:"dash_id,omitempty"`
	PublicKey   string                 `json:"public_key"`
	CipherLog   string                 `json:"cipher_log,omitempty"`
	CipherCount string                 `json:"cipher_count,omitempty"`
	PlainLog    string                 `json:"_log,omitempty"`
	*Log        `json:"log,omitempty"` // deprecated field, do not support long messages
	*Count      `json:"count,omitempty"`
	Sig         string `json:"sig,omitempty"`
	Chunk       struct {
		Uid string `json:"uid,omitempty"`
		Ts  int64  `json:"ts,omitempty"`
		I   int    `json:"i,omitempty"`
		N   int    `json:"n,omitempty"`
	} `json:"chunk"`
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

func (lp *LogPackage) Chunkify(n int, priv string) ([][]byte, error) {
	uid := helpers.RandString(6)
	err := lp.Sign(uid, 0, 1, priv)
	if err != nil {
		return nil, err
	}

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

	for i, chunk := range chunks {
		lpi := *lp

		err = lpi.Sign(uid, i, len(result), priv)
		if err != nil {
			return nil, err
		}

		if lp.CipherLog != "" {
			lpi.CipherLog = chunk
		} else {
			lpi.PlainLog = chunk
		}

		result[i], err = json.Marshal(lpi)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (lp *LogPackage) CalcSig(privBase64 string) (signatureBase64 string, err error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privBase64)
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("%d|%s|%d|%d", lp.Chunk.Ts, lp.Chunk.Uid, lp.Chunk.I, lp.Chunk.N)
	signature, err := cipher.EncryptAes([]byte(message), privateKeyBytes)
	if err != nil {
		return "", err
	}
	fmt.Printf("\n____DEBUG %v %v %v %v \n\n", privBase64, message, base64.StdEncoding.EncodeToString(signature), lp.Chunk)
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (lp *LogPackage) Sign(uid string, i int, n int, privBase64 string) error {
	lp.Chunk.Uid = uid
	lp.Chunk.I = i
	lp.Chunk.N = n
	lp.Chunk.Ts = time.Now().Unix()
	signature, err := lp.CalcSig(privBase64)
	if err != nil {
		return err
	}
	lp.Sig = signature
	return nil
}
