package installation

import "context"

// iInstallationProxy reports where Claude Code is installed and which release it is.
type iInstallationProxy interface {
	Inspect(ctx context.Context) (InstallationVo, error)
}
