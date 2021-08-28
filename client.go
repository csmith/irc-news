package ircplugins

import (
	"context"
	"flag"

	"github.com/greboid/irc-bot/v5/plugins"
)

var (
	rpcHost  = flag.String("host", "localhost:8001", "Host and port to connect to RPC server on")
	rpcToken = flag.String("token", "isedjfiuwserfuesd", "Token to use to authenticate RPC requests")
)

// RpcClient is a simple wrapper around the bot's RPC interface.
type RpcClient struct {
	helper *plugins.PluginHelper
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient creates a new RpcClient and connects to the user-supplied host.
func NewClient() (*RpcClient, error) {
	helper, err := plugins.NewHelper(*rpcHost, *rpcToken)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RpcClient{
		helper: helper,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Send sends a message to a channel.
func (r *RpcClient) Send(channel, message string) error {
	return r.helper.SendChannelMessageWithContext(r.ctx, channel, message)
}

// Close disconnects from the RPC service.
func (r *RpcClient) Close() error {
	r.cancel()
	return nil
}
