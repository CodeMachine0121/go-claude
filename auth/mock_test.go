package auth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockSignInProxy struct {
	mock.Mock
}

func (m *mockSignInProxy) ReadStatus(ctx context.Context) (StatusVo, error) {
	args := m.Called(ctx)
	return args.Get(0).(StatusVo), args.Error(1)
}

func (m *mockSignInProxy) SignIn(ctx context.Context, signIn SignInVo) error {
	args := m.Called(ctx, signIn)
	return args.Error(0)
}

func (m *mockSignInProxy) SignOut(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
