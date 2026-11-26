//go:build integration

package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/query"
)

const fastestModel = "haiku"

// requireLiveTests skips tests that send requests to Claude, because they use
// the person's usage. Run them with GO_CLAUDE_LIVE_TESTS=1.
func requireLiveTests(t *testing.T) {
	t.Helper()
	if os.Getenv("GO_CLAUDE_LIVE_TESTS") != "1" {
		t.Skip("uses Claude usage; set GO_CLAUDE_LIVE_TESTS=1 to run")
	}
}

func everydayClient(t *testing.T, ctx context.Context) *claude.Client {
	t.Helper()
	requireClaudeCode(t)
	client, err := claude.New(ctx)
	require.NoError(t, err)
	return client
}

func TestAMissingWorkFolderIsRefused(t *testing.T) {
	ctx := contextFor(t, 10*time.Second)
	client := everydayClient(t, ctx)

	_, err := client.Query().Ask(ctx, query.Request{
		Prompt: "hi", WorkingDirectory: filepath.Join(t.TempDir(), "missing"),
	})

	require.ErrorIs(t, err, query.ErrWorkFolderMissing)
}

func TestAnUnknownConversationIsRefused(t *testing.T) {
	ctx := contextFor(t, time.Minute)
	client := everydayClient(t, ctx)

	_, err := client.Query().Ask(ctx, query.Request{
		Prompt: "hi", Model: fastestModel, WorkingDirectory: t.TempDir(),
		SessionID: "00000000-0000-0000-0000-000000000000",
	})

	require.ErrorIs(t, err, query.ErrConversationNotFound)
}

func TestAnUnavailableModelIsReportedWithClaudesExplanation(t *testing.T) {
	ctx := contextFor(t, time.Minute)
	client := everydayClient(t, ctx)

	_, err := client.Query().Ask(ctx, query.Request{Prompt: "hi", Model: "no-such-model", WorkingDirectory: t.TempDir()})

	require.ErrorIs(t, err, query.ErrClaudeReportedProblem)
	assert.Contains(t, err.Error(), "no-such-model")
}

func TestARequestFromSomeoneNotSignedInFails(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := contextFor(t, time.Minute)
	client, err := claude.New(ctx, claude.WithProfile(t.TempDir()))
	require.NoError(t, err)

	_, err = client.Query().Ask(ctx, query.Request{Prompt: "hi", Model: fastestModel, WorkingDirectory: t.TempDir()})

	require.ErrorIs(t, err, query.ErrNotSignedIn)
}

func TestARequestGetsAFinalReply(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)

	result, err := client.Query().Ask(ctx, query.Request{
		Prompt: "Reply with exactly: OK", Model: fastestModel, WorkingDirectory: t.TempDir(),
	})

	require.NoError(t, err)
	assert.Equal(t, "OK", result.Text)
	assert.Equal(t, 1, result.Turns)
	assert.NotEmpty(t, result.SessionID)
	assert.NotEmpty(t, result.CostUSD)
}

func TestTheRoundLimitStopsClaudeEarly(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)

	_, err := client.Query().Ask(ctx, query.Request{
		Prompt:   "Use the Bash tool to run 'ls', then tell me how many entries there are.",
		MaxTurns: 1, Model: fastestModel, AllowedTools: []string{"Bash"}, WorkingDirectory: t.TempDir(),
	})

	var limit *query.RoundLimitError
	require.ErrorAs(t, err, &limit)
	assert.GreaterOrEqual(t, limit.Turns, 1)
	assert.NotEmpty(t, limit.SessionID)
	assert.NotEmpty(t, limit.CostUSD)
}

func TestClaudeWorksInTheChosenFolderWithAllowedTools(t *testing.T) {
	requireLiveTests(t)
	folder := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(folder, "marker-7f3a.txt"), nil, 0o644))
	ctx := contextFor(t, 3*time.Minute)
	client := everydayClient(t, ctx)

	result, err := client.Query().Ask(ctx, query.Request{
		Prompt:   "Use the Bash tool to run 'ls' in the current folder, then reply with only the file name you see.",
		MaxTurns: 5, Model: fastestModel, AllowedTools: []string{"Bash"}, WorkingDirectory: folder,
	})

	require.NoError(t, err)
	assert.Contains(t, result.Text, "marker-7f3a.txt")
	assert.LessOrEqual(t, result.Turns, 5)
}

func TestALaterRequestContinuesTheConversation(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 3*time.Minute)
	client := everydayClient(t, ctx)
	folder := t.TempDir()
	first, err := client.Query().Ask(ctx, query.Request{Prompt: "Reply with exactly: OK", Model: fastestModel, WorkingDirectory: folder})
	require.NoError(t, err)

	second, err := client.Query().Ask(ctx, query.Request{
		Prompt: "What did you just reply? Answer in one word.", Model: fastestModel,
		WorkingDirectory: folder, SessionID: first.SessionID,
	})

	require.NoError(t, err)
	assert.Contains(t, second.Text, "OK")
	assert.Equal(t, first.SessionID, second.SessionID)
}

func TestGivingUpStopsClaudeWhileItWorks(t *testing.T) {
	requireLiveTests(t)
	client := everydayClient(t, contextFor(t, 10*time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	var cancelledAt time.Time
	time.AfterFunc(1500*time.Millisecond, func() { cancelledAt = time.Now(); cancel() })

	_, err := client.Query().Ask(ctx, query.Request{
		Prompt: "Write a 2000-word essay about the history of the printing press.", Model: fastestModel,
		WorkingDirectory: t.TempDir(),
	})

	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(cancelledAt), time.Second)
}
