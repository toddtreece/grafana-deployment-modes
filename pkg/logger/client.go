package logger

import (
	"context"
	"fmt"
	"net"

	"github.com/grafana/dskit/services"
)

var _ services.Service = (*Client)(nil)
var _ Logger = (*Client)(nil)

// Client is a logger that logs to a remote server over TCP.
type Client struct {
	*services.BasicService
	messages chan string
	conn     net.Conn
}

func NewClient() *Client {
	c := &Client{
		messages: make(chan string, 1),
	}
	c.BasicService = services.NewBasicService(c.start, c.run, c.stop)
	return c
}

func (c *Client) start(ctx context.Context) error {
	conn, err := net.Dial("tcp", ADDRESS)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *Client) stop(failure error) error {
	defer c.conn.Close()
	return failure
}

func (c *Client) Log(msg string) {
	if c.State() != services.Running {
		return
	}

	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("error writing to connection: %w", err)

		// attempt to reconnect
		_ = c.start(context.Background())
	}
}
