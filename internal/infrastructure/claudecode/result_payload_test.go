package claudecode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/query"
)

// Recorded from Claude Code 2.1.332 (`claude -p --output-format json`), trimmed
// to the fields that matter plus a few neighbours.
const (
	recordedFinished = `{"type":"result","subtype":"success","is_error":false,"duration_ms":2099,
"num_turns":1,"result":"OK","stop_reason":"end_turn","session_id":"986447c4-86c1-4f1b-9de3-c1762657e76d",
"total_cost_usd":0.00559738,"terminal_reason":"completed","permission_denials":[]}`
	recordedRoundLimit = `{"type":"result","subtype":"error_max_turns","is_error":true,"num_turns":2,
"stop_reason":"tool_use","session_id":"84ef484d-d794-43f1-9ae3-e1e22b73286c",
"total_cost_usd":0.0055135800000000006,"terminal_reason":"max_turns",
"errors":["Reached maximum number of turns (1)"]}`
	recordedUnavailableModel = `{"type":"result","subtype":"success","is_error":true,"num_turns":1,
"result":"There's an issue with the selected model (no-such-model). It may not exist or you may not have access to it. Run --model to pick a different model.",
"session_id":"d0271584-ba37-4ed0-9c5f-a5f5f7647f0e","total_cost_usd":0,"terminal_reason":"api_error","api_error_status":404}`
	recordedInterrupted = `{"type":"result","subtype":"error_during_execution","is_error":true,"num_turns":2,
"total_cost_usd":0.007689830000000001,"session_id":"5794e35b-050e-4c5d-b18f-9958a3886063",
"terminal_reason":"aborted_streaming","errors":["[ede_diagnostic] result_type=user last_content_type=n/a stop_reason=null"]}`
	recordedQueryNotSignedIn = `{"type":"result","subtype":"success","is_error":true,"num_turns":1,
"result":"Not logged in · Please run /login","session_id":"512b579c-92b7-467f-b2e6-f095d4742ad8",
"total_cost_usd":0,"terminal_reason":"api_error","api_error_status":null}`
)

func TestNewResultPayload_ToVo(t *testing.T) {
	tests := []struct {
		name     string
		answer   string
		expected query.OutcomeVo
		cost     string
	}{
		{
			name:   "final reply",
			answer: recordedFinished,
			expected: query.OutcomeVo{
				Kind: query.OutcomeFinished, Text: "OK", Turns: 1,
				SessionID: "986447c4-86c1-4f1b-9de3-c1762657e76d", Explanation: "OK",
			},
			cost: "0.00559738",
		},
		{
			name:   "round limit reached keeps rounds, cost and conversation",
			answer: recordedRoundLimit,
			expected: query.OutcomeVo{
				Kind: query.OutcomeRoundLimitReached, Turns: 2,
				SessionID: "84ef484d-d794-43f1-9ae3-e1e22b73286c", Explanation: "Reached maximum number of turns (1)",
			},
			cost: "0.00551358",
		},
		{
			name:   "unavailable model is a problem with claude's explanation",
			answer: recordedUnavailableModel,
			expected: query.OutcomeVo{
				Kind: query.OutcomeProblem, Turns: 1, SessionID: "d0271584-ba37-4ed0-9c5f-a5f5f7647f0e",
				Text:        "There's an issue with the selected model (no-such-model). It may not exist or you may not have access to it. Run --model to pick a different model.",
				Explanation: "There's an issue with the selected model (no-such-model). It may not exist or you may not have access to it. Run --model to pick a different model.",
			},
			cost: "0",
		},
		{
			name:   "not signed in",
			answer: recordedQueryNotSignedIn,
			expected: query.OutcomeVo{
				Kind: query.OutcomeNotSignedIn, Turns: 1, SessionID: "512b579c-92b7-467f-b2e6-f095d4742ad8",
				Text: "Not logged in · Please run /login", Explanation: "Not logged in · Please run /login",
			},
			cost: "0",
		},
		{
			name: "an unknown earlier conversation",
			answer: `{"type":"result","subtype":"success","is_error":true,"num_turns":0,"total_cost_usd":0,` +
				`"result":"No conversation found with session ID: 00000000-0000-0000-0000-000000000000"}`,
			expected: query.OutcomeVo{
				Kind: query.OutcomeConversationNotFound,
				Text: "No conversation found with session ID: 00000000-0000-0000-0000-000000000000", Explanation: "No conversation found with session ID: 00000000-0000-0000-0000-000000000000",
			},
			cost: "0",
		},
		{
			name:   "an interrupted reply",
			answer: recordedInterrupted,
			expected: query.OutcomeVo{
				Kind: query.OutcomeInterrupted, Turns: 2, SessionID: "5794e35b-050e-4c5d-b18f-9958a3886063",
				Explanation: "[ede_diagnostic] result_type=user last_content_type=n/a stop_reason=null",
			},
			cost: "0.00768983",
		},
		{
			name:     "an error outcome without the error flag is still a problem",
			answer:   `{"type":"result","subtype":"error_during_execution","num_turns":1,"total_cost_usd":0,"errors":["tool crashed"]}`,
			expected: query.OutcomeVo{Kind: query.OutcomeProblem, Turns: 1, Explanation: "tool crashed"},
			cost:     "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := newResultPayload([]byte(tt.answer))
			require.NoError(t, err)
			expectedCost, err := query.NewCostVo(tt.cost)
			require.NoError(t, err)
			tt.expected.Cost = expectedCost

			assert.Equal(t, tt.expected, payload.ToVo(Credential{}))
		})
	}
}

func TestNewResultPayload_Unreadable(t *testing.T) {
	tests := []struct {
		name   string
		answer string
	}{
		{name: "plain text", answer: "No conversation found with session ID: 00000000-0000-0000-0000-000000000000"},
		{name: "not a final result", answer: `{"type":"assistant","total_cost_usd":0}`},
		{name: "cost that is not an amount", answer: `{"type":"result","total_cost_usd":-1}`},
		{name: "missing cost", answer: `{"type":"result","result":"OK"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newResultPayload([]byte(tt.answer))

			assert.Error(t, err)
		})
	}
}

func TestResultPayload_ToVo_HidesASecretThatJSONEscaped(t *testing.T) {
	answer := `{"type":"result","is_error":true,"num_turns":1,"total_cost_usd":0,` +
		`"result":"rejected pass sk-ant-oat01-SE\u0043RET","errors":["sk-ant-oat01-SE\u0043RET is invalid"]}`
	payload, err := newResultPayload([]byte(answer))
	require.NoError(t, err)

	outcome := payload.ToVo(NewOAuthToken("sk-ant-oat01-SECRET"))

	assert.Equal(t, "rejected pass [REDACTED]", outcome.Text)
	assert.Equal(t, "rejected pass [REDACTED]\n[REDACTED] is invalid", outcome.Explanation)
	assert.NotContains(t, outcome.Text+outcome.Explanation, "sk-ant-oat01-SECRET")
}
