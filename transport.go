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

func (conn *Transport) Connect(conf *Config) error {
	var err error
	if conf.Udp != "" {
		conn.Conn, err = net.Dial("udp", conf.Udp)
	} else {
		conn.GrpcConn, err = grpc.Dial(conf.Grpc, grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn.GrpcClient = pb.NewLogRpcClient(conn.GrpcConn)
	}
	conn.Config = conf
	return err
}

func (conn *Transport) Close() error {
	if conn.GrpcConn != nil {
		return conn.GrpcConn.Close()
	} else {
		return conn.Conn.Close()
	}
}

func (conn *Transport) pushGrpc(lp *types.LogPackage) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//fmt.Println(string(lp.ProtoBytes()), len(lp.ProtoBytes()))
	req := lp.Proto()
	_, err := conn.GrpcClient.Push(ctx, req)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (conn *Transport) PushLog(log *types.Log) (int, error) {
	if conn.Conn == nil && conn.GrpcConn == nil {
		return 0, nil
	}

	lp := types.LogPackage{
		DashId:    conn.Config.DashId,
		PublicKey: conn.Config.PublicKey,
		Log:       log,
	}

	if conn.Config.NoCipher == false {
		err := lp.EncryptLog(conn.Config.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	if conn.GrpcConn != nil {
		return conn.pushGrpc(&lp)
	}

	if conn.Config.NoCipher == true {
		err := lp.SerializeLog()
		if err != nil {
			return 0, err
		}
	}

	chunks, err := lp.Chunkify(MAX_MESSAGE_SIZE, conn.Config.PrivateKey)
	if err != nil {
		return 0, err
	}

	for i, chunk := range chunks {
		_, err = conn.Conn.Write(chunk)
		//fmt.Println(err, len(chunk))
		if err != nil {
			return i, err
		}
	}

	return len(chunks), nil
}

func (conn *Transport) PushCount(count *types.Count) (int, error) {
	if conn.Conn == nil && conn.GrpcConn == nil {
		return 0, nil
	}
	lp := types.LogPackage{
		DashId:    conn.Config.DashId,
		PublicKey: conn.Config.PublicKey,
		Count:     count,
	}

	if !conn.Config.NoCipher {
		err := lp.EncryptCount(conn.Config.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	if conn.GrpcConn != nil {
		return conn.pushGrpc(&lp)
	}

	msg, err := gojson.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = conn.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}
