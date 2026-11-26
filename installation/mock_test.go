package installation

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockInstallationProxy struct {
	mock.Mock
}

func (m *mockInstallationProxy) Inspect(ctx context.Context) (InstallationVo, error) {
	args := m.Called(ctx)
	return args.Get(0).(InstallationVo), args.Error(1)
}
