//go:build integration

// Package test runs the library against the real Claude Code CLI.
//
//	go test -tags integration ./test/...
//
// Tests skip when Claude Code is not installed. Tests that need the person's
// everyday sign-in skip when it is signed out. Nothing here changes the
// everyday sign-in: sign-outs only run against isolated profile directories.
package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	claude "github.com/CodeMachine0121/go-claude"
	"github.com/CodeMachine0121/go-claude/auth"
	"github.com/CodeMachine0121/go-claude/installation"
)

func requireClaudeCode(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude code is not installed")
	}
	return path
}

func context10s(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// withoutInheritedCredentials keeps the runner's own credential variables from
// changing what the everyday sign-in reports.
func withoutInheritedCredentials(t *testing.T) {
	t.Helper()
	for _, name := range []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_OAUTH_TOKEN", "CLAUDE_CONFIG_DIR"} {
		if value, ok := os.LookupEnv(name); ok {
			t.Setenv(name, value)
			require.NoError(t, os.Unsetenv(name))
		}
	}
}

// signedOutProfile builds a client on a fresh profile and proves it is isolated
// before any test signs out: if the profile reported a signed-in account, the
// isolation is broken and signing out could end the person's everyday sign-in.
func signedOutProfile(t *testing.T, ctx context.Context, options ...claude.Option) *claude.Client {
	t.Helper()
	withoutInheritedCredentials(t)
	client, err := claude.New(ctx, append([]claude.Option{claude.WithProfile(t.TempDir())}, options...)...)
	require.NoError(t, err)
	status, err := client.Auth().Status(ctx)
	require.NoError(t, err)
	if status.SignedIn {
		t.Fatal("a fresh profile reports a signed-in account: profile isolation is broken, refusing to sign out")
	}
	return client
}

func requireEverydaySignedIn(t *testing.T, ctx context.Context) auth.StatusDto {
	t.Helper()
	client, err := claude.New(ctx)
	require.NoError(t, err)
	status, err := client.Auth().Status(ctx)
	require.NoError(t, err)
	if !status.SignedIn {
		t.Skip("the everyday claude code sign-in is signed out")
	}
	return status
}

func TestClaudeCodeIsFoundAutomatically(t *testing.T) {
	path := requireClaudeCode(t)
	ctx := context10s(t)
	printed, err := exec.CommandContext(ctx, path, "--version").Output()
	require.NoError(t, err)
	expected := strings.Fields(string(printed))[0]

	client, err := claude.New(ctx)

	require.NoError(t, err)
	assert.Equal(t, expected, client.Installation().Version)
	assert.Equal(t, path, client.Installation().Path)
}

func TestTheChosenCopyIsUsed(t *testing.T) {
	path := requireClaudeCode(t)
	chosen := filepath.Join(t.TempDir(), "claude")
	if err := os.Symlink(path, chosen); err != nil {
		t.Skipf("cannot create a second copy: %v", err)
	}

	client, err := claude.New(context10s(t), claude.WithCLIPath(chosen))

	require.NoError(t, err)
	assert.Equal(t, chosen, client.Installation().Path)
}

func TestMissingClaudeCodeIsRefused(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	client, err := claude.New(context10s(t))

	require.ErrorIs(t, err, installation.ErrNotInstalled)
	assert.Contains(t, err.Error(), "npm install -g @anthropic-ai/claude-code")
	assert.Nil(t, client)
}

func TestAChosenCopyThatDoesNotExistIsRefused(t *testing.T) {
	requireClaudeCode(t)

	client, err := claude.New(context10s(t), claude.WithCLIPath(filepath.Join(t.TempDir(), "claude")))

	require.ErrorIs(t, err, installation.ErrNotInstalled)
	assert.Nil(t, client)
}

func TestClaudeCodeDisappearingAfterPreparationIsReported(t *testing.T) {
	path := requireClaudeCode(t)
	ctx := context10s(t)
	chosen := filepath.Join(t.TempDir(), "claude")
	if err := os.Symlink(path, chosen); err != nil {
		t.Skipf("cannot create a second copy: %v", err)
	}
	client := signedOutProfile(t, ctx, claude.WithCLIPath(chosen))
	require.NoError(t, os.Remove(chosen))

	_, statusErr := client.Auth().Status(ctx)
	_, loginErr := client.Auth().Login(ctx, auth.LoginRequest{})
	logoutErr := client.Auth().Logout(ctx)

	assert.ErrorIs(t, statusErr, installation.ErrNotInstalled)
	assert.ErrorIs(t, loginErr, installation.ErrNotInstalled)
	assert.ErrorIs(t, logoutErr, installation.ErrNotInstalled)
}

func TestANewProfileStartsSignedOut(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := context10s(t)
	client, err := claude.New(ctx, claude.WithProfile(t.TempDir()))
	require.NoError(t, err)

	status, err := client.Auth().Status(ctx)

	require.NoError(t, err)
	assert.False(t, status.SignedIn)
	assert.Equal(t, auth.AccountKindNotSignedIn, status.AccountKind)
}

func TestNoProfileMeansTheEverydaySignIn(t *testing.T) {
	path := requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := context10s(t)
	everyday := requireEverydaySignedIn(t, ctx)
	reported, _ := exec.CommandContext(ctx, path, "auth", "status").Output()

	assert.True(t, everyday.SignedIn)
	assert.Contains(t, string(reported), `"loggedIn": true`)
}

func TestAProfileIsSeparateFromTheEverydaySignIn(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := context10s(t)
	requireEverydaySignedIn(t, ctx)
	profile, err := claude.New(ctx, claude.WithProfile(t.TempDir()))
	require.NoError(t, err)

	status, err := profile.Auth().Status(ctx)

	require.NoError(t, err)
	assert.False(t, status.SignedIn)
	assert.True(t, requireEverydaySignedIn(t, ctx).SignedIn)
}

func TestSigningOutOfAProfileLeavesTheEverydaySignInAlone(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := context10s(t)
	requireEverydaySignedIn(t, ctx)
	profile := signedOutProfile(t, ctx)

	err := profile.Auth().Logout(ctx)

	require.NoError(t, err)
	status, err := profile.Auth().Status(ctx)
	require.NoError(t, err)
	assert.False(t, status.SignedIn)
	assert.True(t, requireEverydaySignedIn(t, ctx).SignedIn)
}

func TestSignOutWhenAlreadySignedOutSucceeds(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	ctx := context10s(t)
	profile := signedOutProfile(t, ctx)

	err := profile.Auth().Logout(ctx)

	require.NoError(t, err)
}

func TestARelativeProfileDirectoryIsUsed(t *testing.T) {
	requireClaudeCode(t)
	withoutInheritedCredentials(t)
	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(previous) })
	ctx := context10s(t)
	profile, err := claude.New(ctx, claude.WithProfile("profile"))
	require.NoError(t, err)

	status, err := profile.Auth().Status(ctx)

	require.NoError(t, err)
	assert.False(t, status.SignedIn)
	assert.DirExists(t, "profile")
}

func TestSuppliedCredentialsAreUsedWithoutSigningIn(t *testing.T) {
	tests := []struct {
		name         string
		option       claude.Option
		expectedKind auth.AccountKind
	}{
		{name: "subscription pass", option: claude.WithOAuthToken("sk-ant-oat01-not-a-real-pass"), expectedKind: auth.AccountKindSubscriptionPass},
		{name: "developer account key", option: claude.WithAPIKey("sk-ant-api03-not-a-real-key"), expectedKind: auth.AccountKindDeveloperAccountKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireClaudeCode(t)
			withoutInheritedCredentials(t)
			ctx := context10s(t)
			client, err := claude.New(ctx, claude.WithProfile(t.TempDir()), tt.option)
			require.NoError(t, err)

			status, err := client.Auth().Status(ctx)

			require.NoError(t, err)
			assert.True(t, status.SignedIn)
			assert.Equal(t, tt.expectedKind, status.AccountKind)
		})
	}
}

func TestASuppliedCredentialReplacesAnInheritedOne(t *testing.T) {
	requireClaudeCode(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-api03-inherited")
	ctx := context10s(t)
	client, err := claude.New(ctx, claude.WithProfile(t.TempDir()), claude.WithOAuthToken("sk-ant-oat01-not-a-real-pass"))
	require.NoError(t, err)

	status, err := client.Auth().Status(ctx)

	require.NoError(t, err)
	assert.Equal(t, auth.AccountKindSubscriptionPass, status.AccountKind)
}

func TestGivingUpStopsTheCheck(t *testing.T) {
	requireClaudeCode(t)
	client, err := claude.New(context10s(t), claude.WithProfile(t.TempDir()))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	_, err = client.Auth().Status(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(started), time.Second)
}

func TestGivingUpStopsTheSignIn(t *testing.T) {
	requireClaudeCode(t)
	client, err := claude.New(context10s(t), claude.WithProfile(t.TempDir()))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Auth().Login(ctx, auth.LoginRequest{})

	require.ErrorIs(t, err, context.Canceled)
}

func TestGivingUpStopsTheSignOut(t *testing.T) {
	requireClaudeCode(t)
	client := signedOutProfile(t, context10s(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Auth().Logout(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestGivingUpStopsPreparation(t *testing.T) {
	requireClaudeCode(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := claude.New(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, client)
}
