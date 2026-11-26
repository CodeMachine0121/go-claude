package auth

import "context"

// UseCase reads, starts and ends the Claude Code sign-in. It is safe for concurrent use.
type UseCase struct {
	proxy iSignInProxy
}

// NewUseCase builds the use case on top of the infrastructure that runs Claude Code.
func NewUseCase(proxy iSignInProxy) *UseCase {
	return &UseCase{proxy: proxy}
}

// Status reports the current sign-in state. Not being signed in is reported as
// StatusDto{SignedIn: false}, not as an error.
func (u *UseCase) Status(ctx context.Context) (StatusDto, error) {
	observed, err := u.proxy.ReadStatus(ctx)
	if err != nil {
		return StatusDto{}, err
	}
	return observed.ToDomain().ToDto(), nil
}

// Login lets the person sign in, guided by Claude Code in their terminal and
// browser, and returns the resulting state. It blocks until the person finishes,
// gives up (ErrSignInNotCompleted) or ctx ends.
func (u *UseCase) Login(ctx context.Context, request LoginRequest) (StatusDto, error) {
	signIn, err := NewSignInDomain(request)
	if err != nil {
		return StatusDto{}, err
	}
	if err := u.proxy.SignIn(ctx, signIn.ToVo()); err != nil {
		return StatusDto{}, err
	}
	status, err := u.Status(ctx)
	if err != nil {
		return StatusDto{}, err
	}
	if !status.SignedIn {
		return StatusDto{}, ErrSignInNotCompleted
	}
	return status, nil
}

// Logout signs the person out. Signing out when already signed out succeeds.
func (u *UseCase) Logout(ctx context.Context) error {
	return u.proxy.SignOut(ctx)
}
