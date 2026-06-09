package cmd

import (
	"fmt"
	"os"

	"github.com/tab-dance/tabdance-test/internal/client"
)

func newClient() *client.Client {
	if err := client.EnsureDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return client.New()
}

func parseKeyValue(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], nil
		}
	}
	return s, "", nil
}
