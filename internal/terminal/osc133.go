package terminal

import (
	"strconv"
	"strings"
	"time"
)

// osc133Scanner detects OSC 133 semantic prompt markers (FinalTerm
// convention) in a byte stream that may split sequences across reads:
//
//	ESC ] 133 ; A  — prompt start
//	ESC ] 133 ; B  — command start
//	ESC ] 133 ; C  — output start
//	ESC ] 133 ; D[;code] — command finished (with optional exit status)
//
// Only the events tests care about are tracked: prompt redraws (A) and
// command completion (D), the latter carrying the shell's exit status.
type osc133Scanner struct {
	pending []byte
}

type osc133Events struct {
	Prompt         bool // saw a prompt-start marker (A)
	CommandEnd     bool // saw a command-finished marker (D)
	CommandExit    int  // exit code from D when present
	HadCommandExit bool
}

func (s *osc133Scanner) feed(data []byte) osc133Events {
	var ev osc133Events
	s.pending = append(s.pending, data...)

	for {
		idx := strings.Index(string(s.pending), "\x1b]133;")
		if idx < 0 {
			// Keep the last few bytes: a marker may be split across reads.
			const keep = len("\x1b]133;") - 1
			if len(s.pending) > keep {
				s.pending = append([]byte{}, s.pending[len(s.pending)-keep:]...)
			}
			break
		}

		paramStart := idx + len("\x1b]133;")
		// Find the terminator: BEL or ESC-\\.
		end := -1
		for j := paramStart; j < len(s.pending); j++ {
			c := s.pending[j]
			if c == 0x07 {
				end = j
				break
			}
			if c == 0x1b && j+1 < len(s.pending) && s.pending[j+1] == '\\' {
				end = j
				break
			}
		}
		if end < 0 {
			// Sequence incomplete: retain from the marker onward.
			s.pending = append([]byte{}, s.pending[idx:]...)
			break
		}

		params := string(s.pending[paramStart:end])
		s.pending = append([]byte{}, s.pending[end+1:]...)
		if strings.HasPrefix(params, "D") {
			if code, ok := osc133Exit(params); ok {
				ev.CommandEnd = true
				ev.CommandExit = code
				ev.HadCommandExit = true
			} else {
				ev.CommandEnd = true
			}
		} else if params == "A" || strings.HasPrefix(params, "A;") {
			ev.Prompt = true
		}
	}
	return ev
}

func osc133Exit(params string) (int, bool) {
	parts := strings.Split(params, ";")
	if len(parts) < 2 {
		return 0, false
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	return code, true
}

// osc133State holds the recognized events for a terminal.
type osc133State struct {
	prompt         time.Time
	commandEnd     time.Time
	commandExit    int
	hadCommandExit bool
}
