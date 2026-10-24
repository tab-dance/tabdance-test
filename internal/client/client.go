package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/tab-dance/tabdance-test/internal/daemon"
)

type Client struct {
	socketPath string
}

func New() *Client {
	return &Client{
		socketPath: daemon.SocketPath(),
	}
}

func (c *Client) connect() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to daemon: %w", err)
	}
	return conn, nil
}

// callTimeout bounds how long a single request may take. Tape playback and
// waits can legitimately run for minutes, so this is a hang-guard, not a
// tight timeout.
const callTimeout = 10 * time.Minute

func (c *Client) Call(method string, params map[string]any) (*daemon.Response, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := daemon.Request{Method: method, Params: params}
	enc := json.NewEncoder(conn)
	if deadline := time.Now().Add(callTimeout); true {
		conn.SetWriteDeadline(deadline)
	}
	if err := enc.Encode(&req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	dec := json.NewDecoder(conn)
	var resp daemon.Response
	conn.SetReadDeadline(time.Now().Add(callTimeout))
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &resp, nil
}
