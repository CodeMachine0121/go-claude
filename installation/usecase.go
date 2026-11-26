package installation

import "context"

// UseCase verifies the Claude Code installation. It is safe for concurrent use.
type UseCase struct {
	proxy iInstallationProxy
}

// NewUseCase builds the use case on top of the infrastructure that inspects Claude Code.
func NewUseCase(proxy iInstallationProxy) *UseCase {
	return &UseCase{proxy: proxy}
}

// Verify finds Claude Code and refuses it when it is missing (ErrNotInstalled) or
// older than MinimumSupportedVersion (ErrUnsupportedVersion).
func (u *UseCase) Verify(ctx context.Context) (InstallationDto, error) {
	observed, err := u.proxy.Inspect(ctx)
	if err != nil {
		return InstallationDto{}, err
	}
	installation := observed.ToDomain()
	if err := installation.Verify(); err != nil {
		return InstallationDto{}, err
	}
	return installation.ToDto(), nil
}
