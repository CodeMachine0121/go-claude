package claudecode

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/query"
)

// Lines recorded from Claude Code 2.1.332 (`--output-format stream-json
// --verbose`), trimmed to the fields that matter plus a few neighbours.
const (
	recordedAllowedNotice = `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed","resetsAt":1795779600,` +
		`"rateLimitType":"five_hour","overageStatus":"rejected","isUsingOverage":false},` +
		`"uuid":"7f7133d4-173d-4041-a5ee-b8b7aabfa2fd","session_id":"fa202e22-6ee9-459f-800b-84deac8ab36d"}`
	recordedSystemLine = `{"type":"system","subtype":"init","session_id":"fa202e22-6ee9-459f-800b-84deac8ab36d","tools":["Bash","Read"]}`
)

// A usage-limit refusal cannot be produced on demand, so these lines follow the
// answer schema Claude Code 2.1.332 documents for it: a rejected rate_limit_event,
// an assistant message wrapping the refusal with api_error_params.rate_limit_info,
// and a final answer marked with api_error "usage_limit_reached".
const (
	usageLimitNotice = `{"type":"rate_limit_event","rate_limit_info":{"status":"rejected","resetsAt":1795791600,` +
		`"rateLimitType":"five_hour","isUsingOverage":false},"uuid":"u-1","session_id":"s-limit"}`
	usageLimitRefusal = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text",` +
		`"text":"You've hit your session limit · resets 3pm (UTC)"}]},"is_api_error_message":true,` +
		`"api_error_status":429,"error":"rate_limit","api_error":"usage_limit_reached",` +
		`"api_error_params":{"rate_limit_info":{"status":"rejected","resetsAt":1795791600,"rateLimitType":"five_hour"}},` +
		`"session_id":"s-limit"}`
	usageLimitResult = `{"type":"result","subtype":"success","is_error":true,"api_error_status":429,` +
		`"api_error":"usage_limit_reached","num_turns":1,"result":"You've hit your session limit · resets 3pm (UTC)",` +
		`"session_id":"s-limit","total_cost_usd":0,"terminal_reason":"api_error"}`
)

// usageLimitResetsAt is 1795791600 in unix seconds.
var usageLimitResetsAt = time.Date(2026, time.November, 27, 15, 0, 0, 0, time.UTC)

// followed feeds text to a stream in writes of the given size and returns what it delivered and printed.
func followed(t *testing.T, text string, writeSize int, credential Credential) ([]answer, string) {
	t.Helper()
	var delivered []answer
	printed := &diagnostics{}
	stream := newAnswerStream(credential, printed, func(final answer) { delivered = append(delivered, final) })
	// The recorded final answers are spread over several lines for reading; on the wire each is one line.
	text = strings.NewReplacer(",\n", ",", "{\n", "{").Replace(text)
	for start := 0; start < len(text); start += writeSize {
		written, err := stream.Write([]byte(text[start:min(start+writeSize, len(text))]))
		require.NoError(t, err)
		require.Equal(t, min(writeSize, len(text)-start), written)
	}
	stream.Flush()
	return delivered, printed.String()
}

func TestAnswerStream_UsageLimit(t *testing.T) {
	tests := []struct {
		name             string
		lines            []string
		expectedKind     query.OutcomeKind
		expectedResetsAt time.Time
	}{
		{
			name:             "a refusal carries the reset moment claude code stated",
			lines:            []string{recordedSystemLine, usageLimitNotice, usageLimitRefusal, usageLimitResult},
			expectedKind:     query.OutcomeUsageLimitReached,
			expectedResetsAt: usageLimitResetsAt,
		},
		{
			name:             "the refusal's own record is enough",
			lines:            []string{usageLimitRefusal, usageLimitResult},
			expectedKind:     query.OutcomeUsageLimitReached,
			expectedResetsAt: usageLimitResetsAt,
		},
		{
			name:             "a rejected notice alone is enough",
			lines:            []string{usageLimitNotice, usageLimitResult},
			expectedKind:     query.OutcomeUsageLimitReached,
			expectedResetsAt: usageLimitResetsAt,
		},
		{
			name:         "no reset moment is invented when none was stated",
			lines:        []string{recordedSystemLine, usageLimitResult},
			expectedKind: query.OutcomeUsageLimitReached,
		},
		{
			name:         "a later allowed notice withdraws an earlier reset moment",
			lines:        []string{usageLimitNotice, recordedAllowedNotice, usageLimitResult},
			expectedKind: query.OutcomeUsageLimitReached,
		},
		{
			name:         "usage close to the limit is not a refusal",
			lines:        []string{strings.Replace(recordedAllowedNotice, `"allowed"`, `"allowed_warning"`, 1), recordedFinished},
			expectedKind: query.OutcomeFinished,
		},
		{
			name:         "a rejected notice does not turn a finished reply into a refusal",
			lines:        []string{usageLimitNotice, recordedFinished},
			expectedKind: query.OutcomeFinished,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivered, _ := followed(t, strings.Join(tt.lines, "\n")+"\n", 4096, Credential{})

			require.Len(t, delivered, 1)
			require.NoError(t, delivered[0].err)
			assert.Equal(t, tt.expectedKind, delivered[0].outcome.Kind)
			assert.Equal(t, tt.expectedResetsAt, delivered[0].outcome.ResetsAt)
		})
	}
}

func TestAnswerStream_Lines(t *testing.T) {
	t.Run("lines split across writes are put back together", func(t *testing.T) {
		delivered, _ := followed(t, usageLimitNotice+"\n"+recordedFinished+"\n", 7, Credential{})

		require.Len(t, delivered, 1)
		assert.Equal(t, "OK", delivered[0].outcome.Text)
	})

	t.Run("a last answer without a line break still arrives", func(t *testing.T) {
		delivered, _ := followed(t, recordedSystemLine+"\n"+recordedFinished, 4096, Credential{})

		require.Len(t, delivered, 1)
		assert.Equal(t, query.OutcomeFinished, delivered[0].outcome.Kind)
	})

	t.Run("every final answer of a conversation arrives in order", func(t *testing.T) {
		delivered, _ := followed(t, recordedFinished+"\n\n"+recordedRoundLimit+"\n", 4096, Credential{})

		require.Len(t, delivered, 2)
		assert.Equal(t, query.OutcomeFinished, delivered[0].outcome.Kind)
		assert.Equal(t, query.OutcomeRoundLimitReached, delivered[1].outcome.Kind)
	})

	t.Run("plain text is kept to explain a failure; other lines are not", func(t *testing.T) {
		delivered, printed := followed(t, "No conversation found with session ID: x\n"+recordedSystemLine+"\n", 4096, Credential{})

		assert.Empty(t, delivered)
		assert.Equal(t, "No conversation found with session ID: x\n", printed)
	})

	t.Run("a final answer reaches the developer with secrets hidden", func(t *testing.T) {
		delivered, _ := followed(t,
			`{"type":"result","subtype":"success","num_turns":1,"result":"pass sk-ant-oat01-SECRET","total_cost_usd":0.01}`+"\n",
			4096, NewOAuthToken("sk-ant-oat01-SECRET"))

		require.Len(t, delivered, 1)
		require.NoError(t, delivered[0].err)
		assert.Equal(t, "pass [REDACTED]", delivered[0].outcome.Text)
	})

	t.Run("a final answer that cannot be understood is reported, never dropped", func(t *testing.T) {
		delivered, _ := followed(t, `{"type":"result","result":"sk-ant-oat01-SECRET"}`+"\n", 4096, NewOAuthToken("sk-ant-oat01-SECRET"))

		require.Len(t, delivered, 1)
		require.ErrorIs(t, delivered[0].err, query.ErrUnreadableReply)
		assert.NotContains(t, delivered[0].err.Error(), "sk-ant-oat01-SECRET")
	})
}

func TestDiagnostics_KeepsOnlyTheBeginning(t *testing.T) {
	kept := &diagnostics{}

	written, err := kept.Write([]byte(strings.Repeat("a", keptDiagnostics-2) + "bcdef"))
	_, _ = kept.Write([]byte("more"))

	require.NoError(t, err)
	assert.Equal(t, keptDiagnostics+3, written)
	assert.Equal(t, strings.Repeat("a", keptDiagnostics-2)+"bc", kept.String())
}
