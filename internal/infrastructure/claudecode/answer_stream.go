package claudecode

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/CodeMachine0121/go-claude/query"
)

// keptDiagnostics bounds the text kept to explain why Claude Code stopped.
const keptDiagnostics = 4096

// answer is one final answer: decoded, or why it could not be.
type answer struct {
	outcome query.OutcomeVo
	err     error
}

// diagnostics keeps the beginning of what Claude Code printed outside its
// answers, up to keptDiagnostics bytes, and drops the rest.
type diagnostics struct {
	mutex sync.Mutex
	kept  bytes.Buffer
}

// Write keeps what still fits and reports everything as written.
func (d *diagnostics) Write(text []byte) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	room := keptDiagnostics - d.kept.Len()
	if room > 0 {
		d.kept.Write(text[:min(room, len(text))])
	}
	return len(text), nil
}

// String is what was kept.
func (d *diagnostics) String() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.kept.String()
}

// rateLimitInfo is Claude Code's account of the subscription's usage limit.
// Status is "allowed", "allowed_warning" or "rejected"; ResetsAt is in unix seconds.
type rateLimitInfo struct {
	Status   string `json:"status"`
	ResetsAt int64  `json:"resetsAt"`
}

// errorParameters are the details Claude Code adds to a reply that wraps a
// refusal; a usage-limit refusal says which limit and when it resets.
type errorParameters struct {
	RateLimitInfo *rateLimitInfo `json:"rate_limit_info"`
}

// streamLine is what matters in one answer line before the final answer: a
// `rate_limit_event` carries rate_limit_info, an assistant message wrapping a
// refusal carries api_error_params.
type streamLine struct {
	Type            string           `json:"type"`
	RateLimitInfo   *rateLimitInfo   `json:"rate_limit_info"`
	ErrorParameters *errorParameters `json:"api_error_params"`
}

// answerStream reads Claude Code's answers as they are written, one JSON line
// at a time: every final answer goes to deliver, text that is not JSON goes to
// printed, and everything else is only followed for what it says about the
// usage limit. Nothing else is kept, so a long piece of work costs no memory.
// A final answer refused by the usage limit carries the reset moment last
// stated. Write is called by one writer at a time.
type answerStream struct {
	credential Credential
	printed    *diagnostics
	deliver    func(answer)
	pending    []byte
	resetsAt   time.Time
}

func newAnswerStream(credential Credential, printed *diagnostics, deliver func(answer)) *answerStream {
	return &answerStream{credential: credential, printed: printed, deliver: deliver}
}

// Write follows every complete line and keeps an unfinished one for later.
func (s *answerStream) Write(text []byte) (int, error) {
	s.pending = append(s.pending, text...)
	for {
		end := bytes.IndexByte(s.pending, '\n')
		if end < 0 {
			break
		}
		s.follow(s.pending[:end+1])
		s.pending = s.pending[end+1:]
	}
	if len(s.pending) == 0 {
		s.pending = nil
	}
	return len(text), nil
}

// Flush follows a last line that did not end with a line break.
func (s *answerStream) Flush() {
	s.follow(s.pending)
	s.pending = nil
}

func (s *answerStream) follow(line []byte) {
	if len(bytes.TrimSpace(line)) == 0 {
		return
	}
	var decoded streamLine
	if json.Unmarshal(line, &decoded) != nil {
		_, _ = s.printed.Write(line)
		return
	}
	if decoded.Type == "result" {
		payload, err := newResultPayload(line)
		if err != nil {
			s.deliver(answer{err: &query.UnreadableReplyError{Output: s.credential.Redact(strings.TrimSpace(string(line)))}})
			return
		}
		payload.resetsAt = s.resetsAt
		s.deliver(answer{outcome: payload.ToVo(s.credential)})
		return
	}
	info := decoded.RateLimitInfo
	if decoded.ErrorParameters != nil && decoded.ErrorParameters.RateLimitInfo != nil {
		info = decoded.ErrorParameters.RateLimitInfo
	}
	switch {
	case info == nil:
	case info.Status == "rejected" && info.ResetsAt > 0:
		s.resetsAt = time.Unix(info.ResetsAt, 0).UTC()
	case info.Status == "allowed" || info.Status == "allowed_warning":
		s.resetsAt = time.Time{}
	}
}
