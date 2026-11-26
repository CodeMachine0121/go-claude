package query

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func openConversation(t *testing.T, link *mockConversationLink, request ConversationRequest) *Conversation {
	t.Helper()
	proxy := new(mockQueryProxy)
	proxy.On("Open", mock.Anything, mock.Anything).Return(link, nil)
	conversation, err := NewUseCase(proxy).Open(context.Background(), request)
	require.NoError(t, err)
	return conversation
}

func TestUseCase_Open(t *testing.T) {
	t.Run("every choice reaches claude code", func(t *testing.T) {
		link := new(mockConversationLink)
		proxy := new(mockQueryProxy)
		proxy.On("Open", mock.Anything, QueryVo{
			MaxTurns: 4, Model: "haiku", WorkingDirectory: "/repo", AllowedTools: []string{"Read"}, SessionID: "session-1",
		}).Return(link, nil)

		conversation, err := NewUseCase(proxy).Open(context.Background(), ConversationRequest{
			MaxTurns: 4, Model: "haiku", WorkingDirectory: "/repo", AllowedTools: []string{"Read"}, SessionID: "session-1",
		})

		require.NoError(t, err)
		assert.Equal(t, "session-1", conversation.SessionID())
		assert.Equal(t, "0", conversation.TotalCostUSD())
		proxy.AssertExpectations(t)
	})

	t.Run("a negative round cap is refused before anything starts", func(t *testing.T) {
		proxy := new(mockQueryProxy)

		conversation, err := NewUseCase(proxy).Open(context.Background(), ConversationRequest{MaxTurns: -1})

		require.ErrorIs(t, err, ErrNegativeMaxTurns)
		assert.Equal(t, "the round limit cannot be negative", err.Error())
		assert.Nil(t, conversation)
		proxy.AssertNotCalled(t, "Open", mock.Anything, mock.Anything)
	})

	t.Run("a work folder that does not exist is refused", func(t *testing.T) {
		proxy := new(mockQueryProxy)
		proxy.On("Open", mock.Anything, QueryVo{WorkingDirectory: "/missing"}).Return(nil, ErrWorkFolderMissing)

		conversation, err := NewUseCase(proxy).Open(context.Background(), ConversationRequest{WorkingDirectory: "/missing"})

		require.ErrorIs(t, err, ErrWorkFolderMissing)
		assert.Nil(t, conversation)
	})
}

func TestConversation_RepliesAndCosts(t *testing.T) {
	link := new(mockConversationLink)
	link.On("Send", mock.Anything, "Reply with exactly: ONE").Return(OutcomeVo{
		Kind: OutcomeFinished, Text: "ONE", Turns: 1, SessionID: "session-1", Cost: cost(t, "0.0038563"),
	}, nil).Once()
	link.On("Send", mock.Anything, "What did you just reply? One word.").Return(OutcomeVo{
		Kind: OutcomeFinished, Text: "ONE", Turns: 1, SessionID: "session-1", Cost: cost(t, "0.00612419"),
	}, nil).Once()
	conversation := openConversation(t, link, ConversationRequest{})

	first, err := conversation.Send(context.Background(), "Reply with exactly: ONE")
	require.NoError(t, err)
	assert.Equal(t, ResultDto{Text: "ONE", Turns: 1, SessionID: "session-1", CostUSD: "0.0038563", CostNanodollars: 3_856_300}, first)
	assert.Equal(t, "0.0038563", conversation.TotalCostUSD())

	second, err := conversation.Send(context.Background(), "What did you just reply? One word.")
	require.NoError(t, err)
	assert.Equal(t, "ONE", second.Text)
	assert.Equal(t, "0.00226789", second.CostUSD)
	assert.Equal(t, "0.00612419", conversation.TotalCostUSD())
	assert.Equal(t, "session-1", conversation.SessionID())
}

func TestConversation_AFailedReplyStillCountsItsCost(t *testing.T) {
	link := new(mockConversationLink)
	link.On("Send", mock.Anything, "Run ls").Return(OutcomeVo{
		Kind: OutcomeRoundLimitReached, Turns: 2, SessionID: "session-1", Cost: cost(t, "0.01"),
	}, nil)
	conversation := openConversation(t, link, ConversationRequest{MaxTurns: 1})

	_, err := conversation.Send(context.Background(), "Run ls")

	var limit *RoundLimitError
	require.ErrorAs(t, err, &limit)
	assert.Equal(t, "0.01", limit.CostUSD)
	assert.Equal(t, "0.01", conversation.TotalCostUSD())
}

func TestConversation_EmptyMessageIsRefused(t *testing.T) {
	link := new(mockConversationLink)
	link.On("Send", mock.Anything, "hi").Return(OutcomeVo{Kind: OutcomeFinished, Text: "hello"}, nil)
	conversation := openConversation(t, link, ConversationRequest{})

	_, err := conversation.Send(context.Background(), "  \n ")
	require.ErrorIs(t, err, ErrEmptyRequest)
	assert.Equal(t, "the request is empty", err.Error())

	reply, err := conversation.Send(context.Background(), "hi")
	require.NoError(t, err)
	assert.Equal(t, "hello", reply.Text)
	link.AssertNumberOfCalls(t, "Send", 1)
}

func TestConversation_ASecondMessageWhileAnsweringIsRefused(t *testing.T) {
	link := new(mockConversationLink)
	answering := make(chan struct{})
	release := make(chan struct{})
	link.On("Send", mock.Anything, "Write an essay").Return(OutcomeVo{Kind: OutcomeFinished, Text: "essay"}, nil).
		Run(func(mock.Arguments) { close(answering); <-release })
	conversation := openConversation(t, link, ConversationRequest{})
	firstReply := make(chan ResultDto, 1)
	go func() {
		reply, _ := conversation.Send(context.Background(), "Write an essay")
		firstReply <- reply
	}()
	<-answering

	_, err := conversation.Send(context.Background(), "Another message")

	require.ErrorIs(t, err, ErrStillAnswering)
	assert.Equal(t, "still answering", err.Error())
	close(release)
	assert.Equal(t, "essay", (<-firstReply).Text)
	link.AssertNotCalled(t, "Send", mock.Anything, "Another message")
}

func TestConversation_Interrupt(t *testing.T) {
	t.Run("interrupting a reply being written abandons it and the conversation continues", func(t *testing.T) {
		link := new(mockConversationLink)
		answering := make(chan struct{})
		interrupted := make(chan struct{})
		link.On("Send", mock.Anything, "Write an essay").Return(OutcomeVo{Kind: OutcomeInterrupted}, nil).
			Run(func(mock.Arguments) { close(answering); <-interrupted }).Once()
		link.On("Interrupt", mock.Anything).Return(nil).Run(func(mock.Arguments) { close(interrupted) })
		link.On("Send", mock.Anything, "Reply with exactly: AFTER").Return(OutcomeVo{Kind: OutcomeFinished, Text: "AFTER"}, nil).Once()
		conversation := openConversation(t, link, ConversationRequest{})
		essay := make(chan error, 1)
		go func() {
			_, err := conversation.Send(context.Background(), "Write an essay")
			essay <- err
		}()
		<-answering

		require.NoError(t, conversation.Interrupt(context.Background()))

		err := <-essay
		require.ErrorIs(t, err, ErrInterrupted)
		assert.Equal(t, "interrupted", err.Error())
		after, err := conversation.Send(context.Background(), "Reply with exactly: AFTER")
		require.NoError(t, err)
		assert.Equal(t, "AFTER", after.Text)
	})

	t.Run("interrupting while idle does nothing", func(t *testing.T) {
		link := new(mockConversationLink)
		conversation := openConversation(t, link, ConversationRequest{})

		err := conversation.Interrupt(context.Background())

		require.NoError(t, err)
		link.AssertNotCalled(t, "Interrupt", mock.Anything)
	})
}

func TestConversation_GivingUpKeepsTheConversation(t *testing.T) {
	link := new(mockConversationLink)
	link.On("Send", mock.Anything, "Write an essay").
		Return(OutcomeVo{}, fmt.Errorf("claude code was given up: %w", context.Canceled)).Once()
	link.On("Send", mock.Anything, "Reply with exactly: AFTER").Return(OutcomeVo{Kind: OutcomeFinished, Text: "AFTER"}, nil).Once()
	conversation := openConversation(t, link, ConversationRequest{})

	_, err := conversation.Send(context.Background(), "Write an essay")
	require.ErrorIs(t, err, context.Canceled)

	after, err := conversation.Send(context.Background(), "Reply with exactly: AFTER")
	require.NoError(t, err)
	assert.Equal(t, "AFTER", after.Text)
}

func TestConversation_Ending(t *testing.T) {
	t.Run("closing ends claude code's work and closing again does nothing", func(t *testing.T) {
		link := new(mockConversationLink)
		link.On("Close").Return(nil).Once()
		conversation := openConversation(t, link, ConversationRequest{})

		require.NoError(t, conversation.Close())
		require.NoError(t, conversation.Close())

		link.AssertNumberOfCalls(t, "Close", 1)
	})

	t.Run("sending after closing is refused", func(t *testing.T) {
		link := new(mockConversationLink)
		link.On("Close").Return(nil)
		conversation := openConversation(t, link, ConversationRequest{})
		require.NoError(t, conversation.Close())

		_, err := conversation.Send(context.Background(), "hi")

		require.ErrorIs(t, err, ErrConversationClosed)
		assert.Equal(t, "the conversation is closed", err.Error())
		link.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("closing while claude answers reports the message as closed", func(t *testing.T) {
		link := new(mockConversationLink)
		answering := make(chan struct{})
		closed := make(chan struct{})
		link.On("Send", mock.Anything, "Write an essay").Return(OutcomeVo{}, ErrConversationEnded).
			Run(func(mock.Arguments) { close(answering); <-closed })
		link.On("Close").Return(nil).Run(func(mock.Arguments) { close(closed) })
		conversation := openConversation(t, link, ConversationRequest{})
		essay := make(chan error, 1)
		go func() {
			_, err := conversation.Send(context.Background(), "Write an essay")
			essay <- err
		}()
		<-answering

		require.NoError(t, conversation.Close())

		require.ErrorIs(t, <-essay, ErrConversationClosed)
	})

	t.Run("claude code stopping on its own ends the conversation for good", func(t *testing.T) {
		link := new(mockConversationLink)
		link.On("Send", mock.Anything, "hi").Return(OutcomeVo{}, ErrConversationEnded).Once()
		conversation := openConversation(t, link, ConversationRequest{})

		_, first := conversation.Send(context.Background(), "hi")
		_, later := conversation.Send(context.Background(), "hello again")

		require.ErrorIs(t, first, ErrConversationEnded)
		require.ErrorIs(t, later, ErrConversationEnded)
		assert.Equal(t, "the conversation ended", later.Error())
		link.AssertNumberOfCalls(t, "Send", 1)
	})

	t.Run("an unknown earlier conversation ends it", func(t *testing.T) {
		link := new(mockConversationLink)
		link.On("Send", mock.Anything, "hi").Return(OutcomeVo{}, ErrConversationNotFound).Once()
		conversation := openConversation(t, link, ConversationRequest{SessionID: "00000000-0000-0000-0000-000000000000"})

		_, first := conversation.Send(context.Background(), "hi")
		_, later := conversation.Send(context.Background(), "hi")

		require.ErrorIs(t, first, ErrConversationNotFound)
		require.ErrorIs(t, later, ErrConversationEnded)
	})
}

func TestConversation_NotSignedInKeepsTheConversationOpen(t *testing.T) {
	link := new(mockConversationLink)
	link.On("Send", mock.Anything, "hi").Return(OutcomeVo{Kind: OutcomeNotSignedIn}, nil)
	conversation := openConversation(t, link, ConversationRequest{})

	_, first := conversation.Send(context.Background(), "hi")
	_, second := conversation.Send(context.Background(), "hi")

	require.ErrorIs(t, first, ErrNotSignedIn)
	require.ErrorIs(t, second, ErrNotSignedIn)
	link.AssertNumberOfCalls(t, "Send", 2)
}

func TestCostVo_Since(t *testing.T) {
	tests := []struct {
		name     string
		total    string
		previous string
		expected string
	}{
		{name: "a later reply costs the increase", total: "0.00612419", previous: "0.0038563", expected: "0.00226789"},
		{name: "the first reply costs the whole total", total: "0.0038563", previous: "0", expected: "0.0038563"},
		{name: "a total that went down counts in full", total: "0.001", previous: "0.005", expected: "0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cost(t, tt.total).Since(cost(t, tt.previous)).String())
		})
	}
}
