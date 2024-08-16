package logr_go_client

import (
	"context"
	pb "github.com/504dev/logr-go-client/protos/gen/go"
	"github.com/504dev/logr-go-client/types"
	gojson "github.com/goccy/go-json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

type Transport struct {
	*Config
	net.Conn
	GrpcConn   *grpc.ClientConn
	GrpcClient pb.LogRpcClient
}

func (tp *Transport) Connect(conf *Config) error {
	var err error
	if conf.Udp != "" {
		tp.Conn, err = net.Dial("udp", conf.Udp)
	} else {
		tp.GrpcConn, err = grpc.Dial(conf.Grpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
		tp.GrpcClient = pb.NewLogRpcClient(tp.GrpcConn)
	}
	if err != nil {
		tp.Conn = nil
		tp.GrpcConn = nil
		tp.GrpcClient = nil
	}
	tp.Config = conf
	return err
}

func (tp *Transport) Close() error {
	if tp.GrpcConn != nil {
		return tp.GrpcConn.Close()
	} else if tp.Conn != nil {
		return tp.Conn.Close()
	} else {
		return nil
	}
}

func (tp *Transport) pushGrpc(lp *types.LogPackage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//fmt.Println(string(lp.ProtoBytes()), len(lp.ProtoBytes()))
	req := lp.Proto()
	_, err := tp.GrpcClient.Push(ctx, req)
	return err
}

func (tp *Transport) PushLog(log *types.Log) (int, error) {
	if tp.Conn == nil && tp.GrpcConn == nil {
		return 0, nil
	}

	lp := types.LogPackage{
		DashId:    tp.Config.DashId,
		PublicKey: tp.Config.PublicKey,
		Log:       log,
	}

	if tp.Config.NoCipher == false {
		err := lp.EncryptLog(tp.Config.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	if tp.GrpcConn != nil {
		return 1, tp.pushGrpc(&lp)
	}

	if tp.Config.NoCipher == true {
		err := lp.SerializeLog()
		if err != nil {
			return 0, err
		}
	}

	chunks, err := lp.Chunkify(MAX_MESSAGE_SIZE, tp.Config.PrivateKey)
	if err != nil {
		return 0, err
	}

	messages, err := chunks.Marshal()
	if err != nil {
		return 0, err
	}

	for i, msg := range messages {
		_, err = tp.Conn.Write(msg)
		//fmt.Println(err, len(chunk))
		if err != nil {
			return i, err
		}
	}

	return len(chunks), nil
}

func (tp *Transport) PushCount(count *types.Count) (int, error) {
	if tp.Conn == nil && tp.GrpcConn == nil {
		return 0, nil
	}
	lp := types.LogPackage{
		DashId:    tp.Config.DashId,
		PublicKey: tp.Config.PublicKey,
		Count:     count,
	}

	if !tp.Config.NoCipher {
		err := lp.EncryptCount(tp.Config.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	if tp.GrpcConn != nil {
		return 0, tp.pushGrpc(&lp)
	}

	msg, err := gojson.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = tp.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}
