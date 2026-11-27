//go:build integration

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CodeMachine0121/go-claude/query"
)

// outsideServiceList writes an MCP configuration file into folder.
func outsideServiceList(t *testing.T, folder, name, content string) string {
	t.Helper()
	path := filepath.Join(folder, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestAMissingOutsideServiceListIsRefused(t *testing.T) {
	ctx := contextFor(t, 10*time.Second)
	client := everydayClient(t, ctx)
	folder := t.TempDir()

	_, askErr := client.Query().Ask(ctx, query.Request{
		Prompt: "hi", WorkingDirectory: folder, MCPConfigurationPaths: []string{"missing.json"},
	})
	conversation, openErr := client.Query().Open(ctx, query.ConversationRequest{
		WorkingDirectory: folder, MCPConfigurationPaths: []string{"missing.json"},
	})

	require.ErrorIs(t, askErr, query.ErrInvalidMCPConfiguration)
	require.ErrorIs(t, openErr, query.ErrInvalidMCPConfiguration)
	assert.Nil(t, conversation)
}

func TestAnUnreadableOutsideServiceListFails(t *testing.T) {
	ctx := contextFor(t, time.Minute)
	client := everydayClient(t, ctx)
	folder := t.TempDir()
	outsideServiceList(t, folder, "broken.json", "not a list")

	_, askErr := client.Query().Ask(ctx, query.Request{
		Prompt: "hi", Model: fastestModel, WorkingDirectory: folder, MCPConfigurationPaths: []string{"broken.json"},
	})
	conversation := openConversation(t, client, query.ConversationRequest{
		Model: fastestModel, WorkingDirectory: folder, MCPConfigurationPaths: []string{"broken.json"},
	})
	_, sendErr := conversation.Send(ctx, "hi")

	require.ErrorIs(t, askErr, query.ErrInvalidMCPConfiguration)
	require.ErrorIs(t, sendErr, query.ErrInvalidMCPConfiguration)
}

func TestRoleInstructionsAreFollowed(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)

	result, err := client.Query().Ask(ctx, query.Request{
		Prompt: "What is the codeword? Reply with only the codeword.", Model: fastestModel, MaxTurns: 1,
		WorkingDirectory: t.TempDir(), AppendSystemPrompt: "The codeword is PELICAN.",
	})

	require.NoError(t, err)
	assert.Contains(t, strings.ToUpper(result.Text), "PELICAN")
}

func TestOnlyTheListedOutsideServicesAndNoForbiddenToolsAreAvailable(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)
	folder := t.TempDir()
	outsideServiceList(t, folder, "none.json", `{"mcpServers":{}}`)

	result, err := client.Query().Ask(ctx, query.Request{
		Prompt: "Look at the names of the tools you can call right now. Reply with exactly three words separated by " +
			"spaces: EDIT if a tool named Edit is among them, otherwise NOEDIT; WRITE if a tool named Write is among " +
			"them, otherwise NOWRITE; MCP if any tool name starts with mcp__, otherwise NOMCP.",
		Model: fastestModel, MaxTurns: 1, WorkingDirectory: folder,
		MCPConfigurationPaths: []string{"none.json"}, StrictMCPConfiguration: true,
		DisallowedTools: []string{"Edit", "Write"},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"NOEDIT", "NOWRITE", "NOMCP"}, strings.Fields(strings.ToUpper(result.Text)))
}

func TestAConversationCarriesTheAgentsChoices(t *testing.T) {
	requireLiveTests(t)
	ctx := contextFor(t, 2*time.Minute)
	client := everydayClient(t, ctx)
	folder := t.TempDir()
	outsideServiceList(t, folder, "none.json", `{"mcpServers":{}}`)
	conversation := openConversation(t, client, query.ConversationRequest{
		Model: fastestModel, MaxTurns: 1, WorkingDirectory: folder,
		MCPConfigurationPaths: []string{"none.json"}, StrictMCPConfiguration: true,
		DisallowedTools: []string{"Edit", "Write"}, AppendSystemPrompt: "The codeword is PELICAN.",
	})

	reply, err := conversation.Send(ctx, "What is the codeword? Reply with only the codeword.")

	require.NoError(t, err)
	assert.Contains(t, strings.ToUpper(reply.Text), "PELICAN")
}
