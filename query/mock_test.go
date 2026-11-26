package query

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockQueryProxy struct {
	mock.Mock
}

func (m *mockQueryProxy) Ask(ctx context.Context, query QueryVo) (OutcomeVo, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(OutcomeVo), args.Error(1)
}
