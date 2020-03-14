package ircplugins

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/greboid/irc/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rpcHost  = flag.String("host", "localhost:8001", "Host and port to connect to RPC server on")
	rpcToken = flag.String("token", "isedjfiuwserfuesd", "Token to use to authenticate RPC requests")
)

// RpcClient is a simple wrapper around the bot's RPC interface.
type RpcClient struct {
	conn   *grpc.ClientConn
	client rpc.IRCPluginClient
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient creates a new RpcClient and connects to the user-supplied host.
func NewClient() (*RpcClient, error) {
	conn, err := grpc.Dial(*rpcHost, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(rpc.CtxWithToken(context.Background(), "bearer", *rpcToken))

	return &RpcClient{
		conn:   conn,
		client: rpc.NewIRCPluginClient(conn),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Send sends a message to a channel.
func (r *RpcClient) Send(channel, message string) error {
	_, err := r.client.SendChannelMessage(r.ctx, &rpc.ChannelMessage{Channel: channel, Message: message})
	return err
}

// Close disconnects from the RPC service.
func (r *RpcClient) Close() error {
	r.cancel()
	return r.conn.Close()
}
