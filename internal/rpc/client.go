package rpc

import (
	"net/rpc"
)

type Client struct {
	*rpc.Client
	isOwner bool
}

func NewClient(socket string, isOwner bool) (*Client, error) {
	// TODO implement
	return &Client{isOwner: isOwner}, nil
}

func (c *Client) IsOwner() bool {
	return c.isOwner
}
