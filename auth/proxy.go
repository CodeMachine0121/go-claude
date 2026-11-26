package auth

import "context"

// iSignInProxy performs Claude Code's sign-in operations.
//
// SignIn returns ErrSignInNotCompleted when the person does not finish, and an
// error matching the context's error when the context ends first.
type iSignInProxy interface {
	ReadStatus(ctx context.Context) (StatusVo, error)
	SignIn(ctx context.Context, signIn SignInVo) error
	SignOut(ctx context.Context) error
}
