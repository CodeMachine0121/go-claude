package claudecode

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/CodeMachine0121/go-claude/query"
)

// resultPayload is the final answer of `claude -p --output-format json`.
type resultPayload struct {
	Type          string      `json:"type"`
	Subtype       string      `json:"subtype"`
	IsError       bool        `json:"is_error"`
	Result        string      `json:"result"`
	Errors        []string    `json:"errors"`
	NumberOfTurns int         `json:"num_turns"`
	SessionID     string      `json:"session_id"`
	TotalCostUSD  json.Number `json:"total_cost_usd"`
	Ending        string      `json:"terminal_reason"`
	// RefusalKind names the cause of a refusal, such as "usage_limit_reached".
	RefusalKind string `json:"api_error"`

	cost query.CostVo
	// resetsAt is when a reached usage limit resets, as stated before this
	// answer; zero when it was not stated.
	resetsAt time.Time
}

// earlierUsageLimitWording is how older Claude Code releases worded a usage
// limit refusal, followed by "|" and the reset moment in unix seconds.
const earlierUsageLimitWording = "Claude AI usage limit reached"

var errNotAResult = errors.New("answer is not a final result")

// newResultPayload decodes the answer, keeping the cost exact.
func newResultPayload(data []byte) (resultPayload, error) {
	var payload resultPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return resultPayload{}, err
	}
	if payload.Type != "result" {
		return resultPayload{}, errNotAResult
	}
	cost, err := query.NewCostVo(payload.TotalCostUSD.String())
	if err != nil {
		return resultPayload{}, err
	}
	payload.cost = cost
	return payload, nil
}

// ToVo normalizes the answer into the query feature's outcome. Texts are
// redacted again after decoding: a secret that JSON escapes (a quote, a
// backslash, a \u sequence) does not match in the raw answer.
func (p resultPayload) ToVo(credential Credential) query.OutcomeVo {
	explanation := strings.TrimSpace(strings.Join(append([]string{p.Result}, p.Errors...), "\n"))
	kind := query.OutcomeFinished
	resetsAt := time.Time{}
	switch {
	case p.Subtype == "error_max_turns":
		kind = query.OutcomeRoundLimitReached
	case p.Ending == "aborted_streaming":
		kind = query.OutcomeInterrupted
	case p.IsError && p.RefusalKind == "usage_limit_reached":
		kind, resetsAt = query.OutcomeUsageLimitReached, p.resetsAt
	case p.IsError && strings.HasPrefix(p.Result, earlierUsageLimitWording):
		kind = query.OutcomeUsageLimitReached
		_, stated, _ := strings.Cut(p.Result, "|")
		if seconds, err := strconv.ParseInt(strings.TrimSpace(stated), 10, 64); err == nil && seconds > 0 {
			resetsAt = time.Unix(seconds, 0).UTC()
		}
	case p.IsError && strings.Contains(explanation, "No conversation found"):
		kind = query.OutcomeConversationNotFound
	case p.IsError && strings.Contains(explanation, "Not logged in"):
		kind = query.OutcomeNotSignedIn
	case p.IsError || strings.HasPrefix(p.Subtype, "error"):
		kind = query.OutcomeProblem
	}
	return query.OutcomeVo{
		Kind:        kind,
		Text:        credential.Redact(p.Result),
		Turns:       p.NumberOfTurns,
		SessionID:   p.SessionID,
		Cost:        p.cost,
		Explanation: credential.Redact(explanation),
		ResetsAt:    resetsAt,
	}
}
