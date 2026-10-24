package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/tabwriter"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/internal/client"
	"github.com/tab-dance/tabdance-test/internal/daemon"
	"github.com/tab-dance/tabdance-test/internal/version"
	libghostty "go.mitchellh.com/libghostty"
)

type checkResult struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the environment for headless terminal testing (spawns nothing)",
	RunE: func(cmd *cobra.Command, args []string) error {
		results := runDoctorChecks()

		if doctorJsonOutput {
			data, _ := json.Marshal(results)
			fmt.Println(string(data))
		} else {
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "CHECK\tSTATUS\tDETAIL\n")
			for _, r := range results {
				status := "OK"
				if !r.OK {
					status = "FAIL"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, status, r.Detail)
			}
			w.Flush()
		}

		for _, r := range results {
			if !r.OK {
				os.Exit(1)
			}
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(doctorCmd)
}

var doctorJsonOutput bool

func runDoctorChecks() []checkResult {
	var results []checkResult

	if runtime.GOOS == "windows" {
		results = append(results, checkResult{Name: "platform", OK: false,
			Detail: "no PTY support on " + runtime.GOOS + " (linux/darwin only)"})
	} else {
		results = append(results, checkResult{Name: "platform", OK: true,
			Detail: runtime.GOOS + "/" + runtime.GOARCH + " with PTY support"})
	}

	if term := os.Getenv("TERM"); term == "" {
		results = append(results, checkResult{Name: "term", OK: false,
			Detail: "TERM is unset; programs may take unexpected code paths"})
	} else {
		results = append(results, checkResult{Name: "term", OK: true, Detail: "TERM=" + term})
	}

	if master, peer, err := pty.Open(); err == nil {
		name := ""
		if fi, statErr := peer.Stat(); statErr == nil {
			name = fi.Name()
		}
		master.Close()
		peer.Close()
		results = append(results, checkResult{Name: "pty", OK: true, Detail: "allocated " + name})
	} else {
		results = append(results, checkResult{Name: "pty", OK: false, Detail: "pty.Open failed: " + err.Error()})
	}

	socketDir := filepath.Dir(daemon.SocketPath())
	if err := os.MkdirAll(socketDir, 0700); err == nil {
		probe := filepath.Join(socketDir, ".doctor-probe")
		if err := os.WriteFile(probe, []byte("ok"), 0600); err == nil {
			os.Remove(probe)
			results = append(results, checkResult{Name: "socket-dir", OK: true, Detail: socketDir + " writable"})
		} else {
			results = append(results, checkResult{Name: "socket-dir", OK: false, Detail: err.Error()})
		}
	} else {
		results = append(results, checkResult{Name: "socket-dir", OK: false, Detail: err.Error()})
	}

	term, err := libghostty.NewTerminal(libghostty.WithSize(80, 24))
	if err != nil {
		results = append(results, checkResult{Name: "vt-engine", OK: false, Detail: err.Error()})
	} else {
		term.Close()
		results = append(results, checkResult{Name: "vt-engine", OK: true,
			Detail: "libghostty terminal created and closed"})
	}

	if client.IsDaemonRunning() {
		results = append(results, checkResult{Name: "daemon", OK: true, Detail: "running"})
	} else {
		results = append(results, checkResult{Name: "daemon", OK: true,
			Detail: "not running (starts automatically on first command)"})
	}

	results = append(results, checkResult{Name: "version", OK: true, Detail: version.Long()})

	return results
}
