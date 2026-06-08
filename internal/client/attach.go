package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/tab-dance/tabdance-test/internal/daemon"
)

func (c *Client) Attach(id string) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	req := daemon.Request{
		Method: "attach",
		Params: map[string]any{"id": id},
	}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(&req); err != nil {
		return fmt.Errorf("send attach request: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		<-sigCh
		detachReq := daemon.Request{Method: "detach", Params: map[string]any{"id": id}}
		enc.Encode(&detachReq)
		conn.Close()
		os.Exit(0)
	}()

	_, err = io.Copy(os.Stdout, conn)
	return err
}
