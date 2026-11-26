package claudecode

import (
	"encoding/json"
	"errors"
	"strings"

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

	cost query.CostVo
}

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
	switch {
	case p.Subtype == "error_max_turns":
		kind = query.OutcomeRoundLimitReached
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
	}
}
