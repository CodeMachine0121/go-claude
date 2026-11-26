//go:build integration

package test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/query"
)

func openConversation(t *testing.T, client *claude.Client, request query.ConversationRequest) *query.Conversation {
	t.Helper()
	if request.Model == "" {
		request.Model = fastestModel
	}
	if request.WorkingDirectory == "" {
		request.WorkingDirectory = t.TempDir()
	}
	conversation, err := client.Query().Open(contextFor(t, 10*time.Second), request)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conversation.Close() })
	return conversation
}

func TestOpeningInAMissingWorkFolderIsRefused(t *testing.T) {
	ctx := contextFor(t, 10*time.Second)
	client := everydayClient(t, ctx)

	_, err := client.Query().Open(ctx, query.ConversationRequest{WorkingDirectory: filepath.Join(t.TempDir(), "missing")})

	require.ErrorIs(t, err, query.ErrWorkFolderMissing)
}

func TestSendingAfterClosingIsRefused(t *testing.T) {
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{})
	require.NoError(t, conversation.Close())

	_, err := conversation.Send(contextFor(t, 10*time.Second), "hi")

	require.ErrorIs(t, err, query.ErrConversationClosed)
	require.NoError(t, conversation.Close())
}

func TestContinuingAnUnknownConversationFails(t *testing.T) {
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{SessionID: "00000000-0000-0000-0000-000000000000"})

	_, err := conversation.Send(contextFor(t, time.Minute), "hi")

	require.ErrorIs(t, err, query.ErrConversationNotFound)
	_, later := conversation.Send(contextFor(t, 10*time.Second), "hi")
	require.ErrorIs(t, later, query.ErrConversationEnded)
}

func TestAConversationFromSomeoneNotSignedInFails(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	client, err := claude.New(contextFor(t, 10*time.Second), claude.WithProfile(t.TempDir()))
	require.NoError(t, err)
	conversation := openConversation(t, client, query.ConversationRequest{})

	_, err = conversation.Send(contextFor(t, time.Minute), "hi")

	require.ErrorIs(t, err, query.ErrNotSignedIn)
}

func TestClaudeRemembersEarlierMessages(t *testing.T) {
	requireLiveTests(t)
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{})
	ctx := contextFor(t, 2*time.Minute)

	first, err := conversation.Send(ctx, "Reply with exactly: ONE")
	require.NoError(t, err)
	second, err := conversation.Send(ctx, "What did you just reply? One word.")
	require.NoError(t, err)

	assert.Equal(t, "ONE", first.Text)
	assert.Contains(t, second.Text, "ONE")
	assert.Equal(t, first.SessionID, second.SessionID)
	assert.Equal(t, first.SessionID, conversation.SessionID())
	assert.Equal(t, first.CostNanodollars+second.CostNanodollars, totalNanodollars(t, conversation))
}

func totalNanodollars(t *testing.T, conversation *query.Conversation) int64 {
	t.Helper()
	total, err := query.NewCostVo(conversation.TotalCostUSD())
	require.NoError(t, err)
	return total.Nanodollars()
}

func TestInterruptingALongReplyKeepsTheConversation(t *testing.T) {
	requireLiveTests(t)
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{})
	ctx := contextFor(t, 2*time.Minute)
	essay := make(chan error, 1)
	go func() {
		_, err := conversation.Send(ctx, "Write a 3000-word essay about the history of bridges.")
		essay <- err
	}()
	time.Sleep(3 * time.Second)

	require.NoError(t, conversation.Interrupt(ctx))

	require.ErrorIs(t, <-essay, query.ErrInterrupted)
	after, err := conversation.Send(ctx, "Reply with exactly: AFTER")
	require.NoError(t, err)
	assert.Equal(t, "AFTER", after.Text)
}

func TestGivingUpOnAMessageKeepsTheConversation(t *testing.T) {
	requireLiveTests(t)
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{})
	essayContext, giveUp := context.WithCancel(context.Background())
	cancelled := make(chan time.Time, 1)
	time.AfterFunc(3*time.Second, func() { cancelled <- time.Now(); giveUp() })

	_, err := conversation.Send(essayContext, "Write a 3000-word essay about the history of bridges.")

	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(<-cancelled), time.Second)
	after, err := conversation.Send(contextFor(t, time.Minute), "Reply with exactly: AFTER")
	require.NoError(t, err)
	assert.Equal(t, "AFTER", after.Text)
}

func TestContinuingAnEarlierConversation(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)
	folder := t.TempDir()
	earlier, err := client.Query().Ask(ctx, query.Request{Prompt: "Reply with exactly: OK", Model: fastestModel, WorkingDirectory: folder})
	require.NoError(t, err)
	conversation := openConversation(t, client, query.ConversationRequest{SessionID: earlier.SessionID, WorkingDirectory: folder})

	reply, err := conversation.Send(ctx, "What did you just reply? One word.")

	require.NoError(t, err)
	assert.Contains(t, reply.Text, "OK")
}

func TestClosingEndsClaudeCodesWork(t *testing.T) {
	requireLiveTests(t)
	if _, err := exec.LookPath("pgrep"); err != nil {
		t.Skip("pgrep is not available")
	}
	marker := "conversation-" + strings.ReplaceAll(t.Name(), "/", "-")
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{AllowedTools: []string{"Read(" + marker + ")"}})
	_, err := conversation.Send(contextFor(t, time.Minute), "Reply with exactly: OK")
	require.NoError(t, err)
	require.True(t, exec.Command("pgrep", "-f", marker).Run() == nil, "claude code should be running while the conversation is open")

	require.NoError(t, conversation.Close())

	assert.Eventually(t, func() bool { return exec.Command("pgrep", "-f", marker).Run() != nil }, time.Second, 50*time.Millisecond,
		"claude code kept running after closing")
	require.NoError(t, conversation.Close())
}

func TestAProblemClaudeExplainsKeepsTheConversationOpen(t *testing.T) {
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{Model: "no-such-model"})

	_, first := conversation.Send(contextFor(t, time.Minute), "hi")
	_, second := conversation.Send(contextFor(t, time.Minute), "hi again")

	require.ErrorIs(t, first, query.ErrClaudeReportedProblem)
	assert.Contains(t, first.Error(), "no-such-model")
	require.ErrorIs(t, second, query.ErrClaudeReportedProblem, "the conversation should still answer")
}

func TestAVeryLongMessageReachesClaude(t *testing.T) {
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{Model: "no-such-model"})
	longMessage := strings.Repeat("All work and no play. ", 20_000) // about 440 KB

	_, err := conversation.Send(contextFor(t, time.Minute), longMessage)

	require.ErrorIs(t, err, query.ErrClaudeReportedProblem, "the message should reach Claude, which then rejects the model")
}

func TestTheRoundCapAppliesToEachMessage(t *testing.T) {
	requireLiveTests(t)
	client := everydayClient(t, contextFor(t, 10*time.Second))
	conversation := openConversation(t, client, query.ConversationRequest{MaxTurns: 1, AllowedTools: []string{"Bash"}})
	ctx := contextFor(t, 2*time.Minute)

	_, err := conversation.Send(ctx, "Use the Bash tool to run 'ls', then tell me how many entries there are.")
	var limit *query.RoundLimitError
	require.ErrorAs(t, err, &limit)

	reply, err := conversation.Send(ctx, "Reply with exactly: AFTER")
	require.NoError(t, err)
	assert.Equal(t, "AFTER", reply.Text)
}
