package types

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/504dev/logr-go-client/cipher"
	"github.com/504dev/logr-go-client/helpers"
	pb "github.com/504dev/logr-go-client/protos/gen/go"
	gojson "github.com/goccy/go-json"
	"github.com/golang/protobuf/proto"
	"time"
)

type ChunkInfo struct {
	Uid string `json:"uid,omitempty"`
	Ts  int64  `json:"ts,omitempty"`
	I   int    `json:"i,omitempty"`
	N   int    `json:"n,omitempty"`
}

func (ch *ChunkInfo) CalcSig(privBase64 string) (signatureBase64 string, err error) {
	if ch.N == 0 || ch.I >= ch.N || len(ch.Uid) < 6 {
		return "", errors.New("bad arguments")
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privBase64)
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("%d|%s|%d|%d", ch.Ts, ch.Uid, ch.I, ch.N)
	signature, err := cipher.EncryptAesIv([]byte(message), privateKeyBytes, []byte(ch.Uid))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

type LogPackage struct {
	DashId      int                    `json:"dash_id,omitempty"`
	PublicKey   string                 `json:"public_key"`
	CipherLog   []byte                 `json:"cipher_log,omitempty"`
	CipherCount []byte                 `json:"cipher_count,omitempty"`
	PlainLog    []byte                 `json:"_log,omitempty"`
	*Log        `json:"log,omitempty"` // field do not support long messages over udp
	*Count      `json:"count,omitempty"`
	Sig         string     `json:"sig,omitempty"`
	Chunk       *ChunkInfo `json:"chunk"`
}

func (lp *LogPackage) SerializeLog() error {
	msg, err := gojson.Marshal(lp.Log)
	if err != nil {
		return err
	}
	lp.PlainLog = msg
	lp.Log = nil
	return nil
}

func (lp *LogPackage) DeserializeLog() error {
	log := Log{}
	err := gojson.Unmarshal(lp.PlainLog, &log)
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

	msg, err := gojson.Marshal(lp)
	if err != nil {
		return nil, err
	}

	if len(msg) <= n {
		return [][]byte{msg}, err
	}

	var data []byte
	if len(lp.CipherLog) > 0 {
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

		if len(lp.CipherLog) > 0 {
			lpi.CipherLog = chunk
		} else {
			lpi.PlainLog = chunk
		}

		result[i], err = gojson.Marshal(lpi)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (lp *LogPackage) Sign(uid string, i int, n int, privBase64 string) error {
	chunkInfo := &ChunkInfo{
		Uid: uid,
		Ts:  time.Now().Unix(),
		I:   i,
		N:   n,
	}
	signature, err := chunkInfo.CalcSig(privBase64)
	if err != nil {
		return err
	}
	lp.Chunk = chunkInfo
	lp.Sig = signature
	return nil
}

func (lp *LogPackage) ProtoBytes() []byte {
	res, _ := proto.Marshal(lp.Proto())
	return res
}

func (lp *LogPackage) FromProto(lrp *pb.LogRpcPackage) {
	*lp = LogPackage{
		DashId:      int(lrp.DashId),
		PublicKey:   lrp.PublicKey,
		CipherLog:   lrp.CipherLog,
		CipherCount: lrp.CipherCount,
		PlainLog:    lrp.PlainLog,
	}
	if lrp.Log != nil {
		lp.Log = &Log{
			DashId:    int(lrp.Log.DashId),
			Pid:       int(lrp.Log.Pid),
			Timestamp: lrp.Log.Timestamp,
			Hostname:  lrp.Log.Hostname,
			Logname:   lrp.Log.Logname,
			Level:     lrp.Log.Level,
			Message:   lrp.Log.Message,
			Version:   lrp.Log.Version,
			Initiator: lrp.Log.Initiator,
		}
	}
	if lp.Count != nil {
		lp.Count = &Count{
			DashId:    int(lrp.Log.DashId),
			Timestamp: lrp.Count.Timestamp,
			Hostname:  lrp.Count.Hostname,
			Version:   lrp.Count.Version,
			Logname:   lrp.Count.Logname,
			Keyname:   lrp.Count.Keyname,
			Metrics:   Metrics{},
		}
		if v := lrp.Count.Inc; v != nil {
			lp.Count.Inc(v.Inc)
		}
		if v := lrp.Count.Max; v != nil {
			lp.Count.Max(v.Max)
		}
		if v := lrp.Count.Min; v != nil {
			lp.Count.Min(v.Min)
		}
		if v := lrp.Count.Avg; v != nil {
			lp.Count.Metrics.Avg = &Avg{v.Sum, int(v.Num)}
		}
		if v := lrp.Count.Per; v != nil {
			lp.Count.Metrics.Per = &Per{v.Taken, v.Total}
		}
		if v := lrp.Count.Time; v != nil {
			lp.Count.Metrics.Time = &Time{v.Duration}
		}
	}
}

func (lp *LogPackage) Proto() *pb.LogRpcPackage {
	res := &pb.LogRpcPackage{
		DashId:      uint32(lp.DashId),
		PublicKey:   lp.PublicKey,
		CipherLog:   lp.CipherLog,
		CipherCount: lp.CipherCount,
		PlainLog:    lp.PlainLog,
	}
	if lp.Log != nil {
		res.Log = &pb.LogRpcPackage_Log{
			DashId:    uint32(lp.Log.DashId),
			Pid:       uint32(lp.Log.Pid),
			Timestamp: lp.Log.Timestamp,
			Hostname:  lp.Log.Hostname,
			Logname:   lp.Log.Logname,
			Level:     lp.Log.Level,
			Message:   lp.Log.Message,
			Version:   lp.Log.Version,
			Initiator: lp.Log.Initiator,
		}
	}
	if lp.Count != nil {
		res.Count = &pb.LogRpcPackage_Count{
			DashId:    uint32(lp.Log.DashId),
			Timestamp: lp.Count.Timestamp,
			Hostname:  lp.Count.Hostname,
			Version:   lp.Count.Version,
			Logname:   lp.Count.Logname,
			Keyname:   lp.Count.Keyname,
		}
		if v := lp.Count.Metrics.Inc; v != nil {
			res.Count.Inc = &pb.LogRpcPackage_Count_Inc{Inc: v.Val}
		}
		if v := lp.Count.Metrics.Max; v != nil {
			res.Count.Max = &pb.LogRpcPackage_Count_Max{Max: v.Val}
		}
		if v := lp.Count.Metrics.Min; v != nil {
			res.Count.Min = &pb.LogRpcPackage_Count_Min{Min: v.Val}
		}
		if v := lp.Count.Metrics.Avg; v != nil {
			res.Count.Avg = &pb.LogRpcPackage_Count_Avg{Sum: v.Sum, Num: uint32(v.Num)}
		}
		if v := lp.Count.Metrics.Per; v != nil {
			res.Count.Per = &pb.LogRpcPackage_Count_Per{Taken: v.Taken, Total: v.Total}
		}
		if v := lp.Count.Metrics.Time; v != nil {
			res.Count.Time = &pb.LogRpcPackage_Count_Time{Duration: v.Duration}
		}
	}
	return res
}
