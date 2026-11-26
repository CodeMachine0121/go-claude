//go:build integration

package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/installation"
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
	cancelled := make(chan time.Time, 1)
	time.AfterFunc(1500*time.Millisecond, func() { cancelled <- time.Now(); cancel() })

	_, err := client.Query().Ask(ctx, query.Request{
		Prompt: "Write a 2000-word essay about the history of the printing press.", Model: fastestModel,
		WorkingDirectory: t.TempDir(),
	})

	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(<-cancelled), time.Second)
}

// running reports whether any process runs command: the command itself or
// the shell Claude Code wraps it in. The request reaches Claude Code on
// standard input, so no process of this test carries the command line.
func running(t *testing.T, command string) bool {
	t.Helper()
	return exec.Command("pgrep", "-f", command).Run() == nil
}

func TestGivingUpAlsoStopsCommandsClaudeStarted(t *testing.T) {
	requireLiveTests(t)
	if runtime.GOOS == "windows" {
		t.Skip("commands Claude starts are stopped together only where process groups exist")
	}
	if _, err := exec.LookPath("pgrep"); err != nil {
		t.Skip("pgrep is not available")
	}
	const command = "sleep 47.31"
	client := everydayClient(t, contextFor(t, 10*time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := client.Query().Ask(ctx, query.Request{
			Prompt:   "Use the Bash tool to run exactly this command and wait for it: " + command,
			MaxTurns: 3, Model: fastestModel, AllowedTools: []string{"Bash"}, WorkingDirectory: t.TempDir(),
		})
		done <- err
	}()
	t.Cleanup(func() { _ = exec.Command("pkill", "-f", command).Run() })
	require.Eventually(t, func() bool { return running(t, command) }, 60*time.Second, 200*time.Millisecond,
		"claude never started the command")

	cancel()

	require.ErrorIs(t, <-done, context.Canceled)
	assert.Eventually(t, func() bool { return !running(t, command) }, time.Second, 50*time.Millisecond,
		"the command claude started kept running after giving up")
}

func TestARequestWhenClaudeCodeDisappearedIsRefused(t *testing.T) {
	path := requireClaudeCode(t)
	ctx := contextFor(t, 10*time.Second)
	chosen := filepath.Join(t.TempDir(), "claude")
	if err := os.Symlink(path, chosen); err != nil {
		t.Skipf("cannot create a second copy: %v", err)
	}
	client, err := claude.New(ctx, claude.WithCLIPath(chosen))
	require.NoError(t, err)
	require.NoError(t, os.Remove(chosen))

	_, err = client.Query().Ask(ctx, query.Request{Prompt: "hi", WorkingDirectory: t.TempDir()})

	require.ErrorIs(t, err, installation.ErrNotInstalled)
}

func TestAVeryLongRequestReachesClaude(t *testing.T) {
	ctx := contextFor(t, time.Minute)
	client := everydayClient(t, ctx)
	longRequest := strings.Repeat("All work and no play. ", 20_000) // about 440 KB, past any command-line limit

	_, err := client.Query().Ask(ctx, query.Request{Prompt: longRequest, Model: "no-such-model", WorkingDirectory: t.TempDir()})

	require.ErrorIs(t, err, query.ErrClaudeReportedProblem, "the request should reach Claude, which then rejects the model")
	assert.Contains(t, err.Error(), "no-such-model")
}
