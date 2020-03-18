package ircplugins

import (
	"github.com/greboid/irc/rpc"
	"strings"
)

// Command represents a command that has been invoked on IRC.
type Command struct {
	client    *RpcClient
	Channel   string
	Arguments string
}

// CommandHandler is a function that deals with a invocation.
type CommandHandler func(command Command)

// ListenForCommands listens to all messages and calls handlers if the message starts with the corresponding string.
func (r *RpcClient) ListenForCommands(handlers map[string]CommandHandler) error {
	c, err := r.client.GetMessages(r.ctx, &rpc.Channel{Name: "*"})
	if err != nil {
		return err
	}

	defer func() {
		_ = c.CloseSend()
	}()

	for {
		message, err := c.Recv()
		if err != nil {
			return err
		}

		for command, handler := range handlers {
			if strings.HasPrefix(strings.ToLower(message.Message), command) {
				go handler(Command{
					client:    r,
					Channel:   message.Channel,
					Arguments: strings.TrimSpace(message.Message[len(command):]),
				})
			}
		}
	}
}

// Reply sends a message back to the channel the command was executed in.
func (c Command) Reply(message string) error {
	return c.client.Send(c.Channel, message)
}
