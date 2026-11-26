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

func (m *mockQueryProxy) Open(ctx context.Context, query QueryVo) (IConversationLink, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(IConversationLink), args.Error(1)
}

type mockConversationLink struct {
	mock.Mock
}

func (m *mockConversationLink) Send(ctx context.Context, message string) (OutcomeVo, error) {
	args := m.Called(ctx, message)
	return args.Get(0).(OutcomeVo), args.Error(1)
}

func (m *mockConversationLink) Interrupt(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockConversationLink) Close() error {
	args := m.Called()
	return args.Error(0)
}
