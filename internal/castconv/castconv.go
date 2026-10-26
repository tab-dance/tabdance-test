// Package castconv converts asciicast v3 recordings into deterministic
// tabdance-test tapes. Input events ("i") become Type/Key/Ctrl/Raw tape
// lines, resize events become Resize, markers and exit-status events
// become comments, and WaitStable lines anchor each input cluster so
// replay waits for output to settle instead of sleeping for fixed
// durations.
package castconv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Event struct {
	Time float64
	Kind string // "o", "i", "m", "r", "x"
	Data string
}

type Cast struct {
	Width, Height int
	Events        []Event
}

// Parse reads a cast file (JSON Lines: asciicast v3 header first, then
// interval-timed events; "#" comment lines are ignored). Event times are
// converted to absolute seconds since the recording start.
func Parse(path string) (*Cast, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &Cast{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lineNo := 0
	headerSeen := false
	abs := 0.0
	for sc.Scan() {
		lineNo++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			continue
		}
		if !headerSeen {
			if line[0] != '{' {
				return nil, fmt.Errorf("cast line %d: expected header, got %q", lineNo, line)
			}
			var header struct {
				Version int `json:"version"`
				Term    struct {
					Cols int `json:"cols"`
					Rows int `json:"rows"`
				} `json:"term"`
			}
			if err := json.Unmarshal(line, &header); err != nil {
				return nil, fmt.Errorf("cast header line %d: %w", lineNo, err)
			}
			if header.Version != 3 {
				return nil, fmt.Errorf("cast header line %d: unsupported asciicast version %d (this tool reads version 3 only)", lineNo, header.Version)
			}
			c.Width, c.Height = header.Term.Cols, header.Term.Rows
			headerSeen = true
			continue
		}
		var ev []any
		if err := json.Unmarshal(line, &ev); err != nil || len(ev) < 3 {
			return nil, fmt.Errorf("cast event line %d: malformed", lineNo)
		}
		interval, _ := ev[0].(float64)
		kind, _ := ev[1].(string)
		data, _ := ev[2].(string)
		abs += interval
		c.Events = append(c.Events, Event{Time: abs, Kind: kind, Data: data})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if !headerSeen {
		return nil, fmt.Errorf("cast has no asciicast header")
	}
	if c.Width == 0 {
		c.Width, c.Height = 80, 24
	}
	return c, nil
}

// ToTape converts the event stream into tape line content.
func (c *Cast) ToTapeCast() string {
	var out []string
	out = append(out,
		fmt.Sprintf("Set Width %d", c.Width),
		fmt.Sprintf("Set Height %d", c.Height),
	)
	if speed := medianTypingSpeed(c.Events); speed >= 0 {
		out = append(out, fmt.Sprintf("Set TypingSpeed %s", speed))
	}
	out = append(out, "WaitStable @500ms")

	var pendingInput []string
	var lastInputTime = -1.0
	var gaps []float64
	flushInput := func() {
		if len(pendingInput) == 0 {
			return
		}
		out = append(out, pendingInput...)
		pendingInput = nil
		out = append(out, "WaitStable @250ms")
	}

	for _, ev := range c.Events {
		switch ev.Kind {
		case "i":
			// Long pauses between keystrokes indicate separate interaction
			// steps; separate them so replay waits for each result.
			if lastInputTime >= 0 {
				delta := ev.Time - lastInputTime
				if delta >= 2.0 && len(pendingInput) > 0 {
					flushInput()
				} else {
					gaps = append(gaps, delta)
				}
			}
			lastInputTime = ev.Time
			pendingInput = append(pendingInput, decodeInput(ev.Data)...)
		case "o":
			// output events are implied by replay; waits are synthesized
		case "m":
			if len(pendingInput) > 0 {
				flushInput()
			}
			out = append(out, fmt.Sprintf("# marker: %s", ev.Data))
		case "x":
			// exit status from the recorded session; replay shells stay
			// alive, so this becomes a comment instead of ExpectExit
			if len(pendingInput) > 0 {
				flushInput()
			}
			out = append(out, fmt.Sprintf("# exit status: %s", ev.Data))
		case "r":
			if cols, rows, ok := parseResize(ev.Data); ok {
				if len(pendingInput) > 0 {
					flushInput()
				}
				out = append(out, fmt.Sprintf("Resize %d %d", cols, rows))
			}
		}
	}
	flushInput()

	// Build the line list preserving cluster separators.
	return assemble(out)
}

func assemble(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}

// medianTypingSpeed estimates the per-character replay delay from the gaps
// between consecutive input events inside a cluster (< 2s apart). It returns
// -1 when the recording has no such gaps (single keystroke or burst).
func medianTypingSpeed(events []Event) time.Duration {
	var gaps []float64
	var last = -1.0
	for _, ev := range events {
		if ev.Kind != "i" {
			continue
		}
		if last >= 0 {
			if delta := ev.Time - last; delta < 2.0 {
				gaps = append(gaps, delta)
			}
		}
		last = ev.Time
	}
	if len(gaps) == 0 {
		return -1
	}
	sort.Float64s(gaps)
	median := gaps[len(gaps)/2]
	ms := int64(math.Round(median * 1000))
	return time.Duration(ms) * time.Millisecond
}

type inputDecoder struct {
	out []string
	cur strings.Builder
}

func (d *inputDecoder) flushTyped() {
	if d.cur.Len() > 0 {
		d.out = append(d.out, "Type "+strconv.Quote(d.cur.String()))
		d.cur.Reset()
	}
}

func (d *inputDecoder) line(raw string) { d.out = append(d.out, raw) }

// decodeInput turns raw injected bytes into tape lines.
func decodeInput(data string) []string {
	d := &inputDecoder{}
	i := 0
	for i < len(data) {
		b := data[i]
		switch {
		case b == 0x1b:
			d.flushTyped()
			consumed, line, ok := decodeEscape(data[i:])
			if !ok || consumed <= 0 {
				d.line(commentRest(data[i:]))
				i = len(data)
				break
			}
			if line != "" {
				d.line(line)
			}
			i += consumed
		case b == '\r':
			d.flushTyped()
			d.line("Enter")
			i++
		case b == '\n':
			// ignore: Enter already emitted for \r; lone \n is an Enter too
			d.flushTyped()
			d.line("Enter")
			i++
		case b == '\t':
			d.flushTyped()
			d.line("Tab")
			i++
		case b == 0x7f || b == 0x08:
			d.flushTyped()
			d.line("Backspace")
			i++
		case b < 0x20:
			if name, ok := ctrlName(b); ok {
				d.flushTyped()
				d.line(name)
				i++
				break
			}
			d.flushTyped()
			d.line(fmt.Sprintf("Raw %q", string(rune(b))))
			i++
		default:
			r, size := utf8.DecodeRuneInString(data[i:])
			if r == utf8.RuneError && size <= 1 {
				d.flushTyped()
				d.line(fmt.Sprintf("Raw %q", data[i:i+1]))
				i++
				break
			}
			d.cur.WriteRune(r)
			i += size
		}
	}
	d.flushTyped()
	return d.out
}

func ctrlName(b byte) (string, bool) {
	switch b {
	case 0x03:
		return "Ctrl+C", true
	case 0x04:
		return "Ctrl+D", true
	case 0x1a:
		return "Ctrl+Z", true
	case 0x0c:
		return "Ctrl+L", true
	case 0x12:
		return "Ctrl+R", true
	case 0x01:
		return "Ctrl+A", true
	case 0x05:
		return "Ctrl+E", true
	case 0x17:
		return "Ctrl+W", true
	case 0x15:
		return "Ctrl+U", true
	case 0x0b:
		return "Ctrl+K", true
	case 0x1b:
		return "Escape", true
	}
	return "", false
}

// decodeEscape maps a leading escape sequence to a tape verb, returning
// bytes consumed, the line, and whether the sequence was understood.
func decodeEscape(s string) (consumed int, line string, ok bool) {
	if len(s) < 2 {
		return 0, "", false
	}
	switch s[1] {
	case '[': // CSI
		end := -1
		for j := 2; j < len(s) && j < 128; j++ {
			c := s[j]
			if (c >= '0' && c <= '9') || c == ';' || c == '<' || c == '?' {
				continue
			}
			end = j
			break
		}
		if end < 0 {
			return 0, "", false
		}
		// Mouse events (M/m) cannot be expressed as tape verbs.
		if s[end] == 'M' || s[end] == 'm' {
			return end + 1, "# skipped: mouse event " + strconv.Quote(s[:end+1]), true
		}
		seq := s[2:end]
		name := ""
		switch {
		case s[end] == 'A':
			name = "Up"
		case s[end] == 'B':
			name = "Down"
		case s[end] == 'C':
			name = "Right"
		case s[end] == 'D':
			name = "Left"
		case s[end] == 'H':
			if seq == "" {
				name = "Home"
			}
		case s[end] == 'F':
			name = "End"
		case s[end] == 'u':
			if seq != "" {
				return decodeCSICodepoint(seq)
			}
		default:
			switch string(s[end]) + seq {
			case "3~":
				name = "Delete"
			case "5~":
				name = "PageUp"
			case "6~":
				name = "PageDown"
			case "1~":
				name = "Home"
			case "4~":
				name = "End"
			}
		}
		if name == "" {
			return end + 1, "# skipped: unrecognized CSI " + strconv.Quote(s[:end+1]), true
		}
		return end + 1, name, true
	case 'O': // SS3
		if len(s) < 3 {
			return 0, "", false
		}
		switch s[2] {
		case 'A':
			return 3, "Up", true
		case 'B':
			return 3, "Down", true
		case 'C':
			return 3, "Right", true
		case 'D':
			return 3, "Left", true
		case 'H':
			return 3, "Home", true
		case 'F':
			return 3, "End", true
		case 'P':
			return 3, "F1", true
		case 'Q':
			return 3, "F2", true
		case 'R':
			return 3, "F3", true
		case 'S':
			return 3, "F4", true
		}
		return 0, "", false
	case ']': // OSC: prompts, titles, clipboard — not input semantics
		return oscLength(s), "", true
	case '(':
		if len(s) >= 3 {
			return 3, "", true // charset designation, ignore
		}
		return 0, "", false
	}
	// Bracketed-paste enable/disable and other terminal replies use CSI
	// sequences we cannot map to input verbs; swallow them so the rest of
	// the stream keeps decoding.
	if n := oscLength(s); n > 0 {
		return n, "", true
	}
	return 0, "", false
}

// decodeCSICodepoint handles CSI-u (kitty-style) input encodings: the
// parameters are "codepoint;modifiers;pressed?".
func decodeCSICodepoint(seq string) (consumed int, line string, ok bool) {
	parts := strings.SplitN(strings.TrimSuffix(seq, "u"), ";", 3)
	code, err := strconv.Atoi(parts[0])
	if err != nil || code <= 0 {
		return 0, "", false
	}
	mods := 0
	if len(parts) > 1 {
		mods, _ = strconv.Atoi(parts[1])
	}
	r := rune(code)
	line = "Type " + strconv.Quote(string(r))
	if mods-1 > 0 {
		// Modifier encodings have no faithful tape spelling; keep the raw.
		line = "Raw " + strconv.Quote(string(r))
	}
	return 2 + len(seq) + 1, line, true
}

func parseResize(s string) (int, int, bool) {
	parts := strings.SplitN(s, "x", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	cols, err1 := strconv.Atoi(parts[0])
	rows, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return cols, rows, true
}

func commentRest(rest string) string {
	rest = strings.NewReplacer("\r", "\n", "\x1b", "ESC-").Replace(rest)
	if rest == "" {
		return "# skipped: unrecognized escape sequence at end of input"
	}
	return "# skipped: unrecognized input: " + strconv.Quote(rest)
}

// oscLength returns the byte length of any OSC sequence, including its
// BEL or ESC-\ terminator; -1 when unterminated.
func oscLength(s string) int {
	for i := 2; i < len(s); i++ {
		if s[i] == 0x07 {
			return i + 1
		}
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
			return i + 2
		}
	}
	return -1
}
