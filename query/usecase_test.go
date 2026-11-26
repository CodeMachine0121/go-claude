package query

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func cost(t *testing.T, text string) CostVo {
	t.Helper()
	value, err := NewCostVo(text)
	require.NoError(t, err)
	return value
}

func TestUseCase_Ask(t *testing.T) {
	tests := []struct {
		name          string
		request       Request
		expectedQuery QueryVo
		outcome       OutcomeVo
		expectedDto   ResultDto
	}{
		{
			name:          "a request gets a final reply",
			request:       Request{Prompt: "Reply with exactly: OK"},
			expectedQuery: QueryVo{Prompt: "Reply with exactly: OK"},
			outcome:       OutcomeVo{Kind: OutcomeFinished, Text: "OK", Turns: 1, SessionID: "session-1", Cost: cost(t, "0.00559738")},
			expectedDto:   ResultDto{Text: "OK", Turns: 1, SessionID: "session-1", CostUSD: "0.00559738", CostNanodollars: 5_597_380},
		},
		{
			name:          "claude finishes within the cap",
			request:       Request{Prompt: "Fix the test", MaxTurns: 5},
			expectedQuery: QueryVo{Prompt: "Fix the test", MaxTurns: 5},
			outcome:       OutcomeVo{Kind: OutcomeFinished, Text: "Fixed", Turns: 3, SessionID: "session-2", Cost: cost(t, "0.1")},
			expectedDto:   ResultDto{Text: "Fixed", Turns: 3, SessionID: "session-2", CostUSD: "0.1", CostNanodollars: 100_000_000},
		},
		{
			name: "every setting reaches claude code",
			request: Request{
				Prompt: "List the files", Model: "haiku", WorkingDirectory: "/repo",
				AllowedTools: []string{"Bash", "Read"}, SessionID: "session-3",
			},
			expectedQuery: QueryVo{
				Prompt: "List the files", Model: "haiku", WorkingDirectory: "/repo",
				AllowedTools: []string{"Bash", "Read"}, SessionID: "session-3",
			},
			outcome:     OutcomeVo{Kind: OutcomeFinished, Text: "a.go", Turns: 2, SessionID: "session-3", Cost: cost(t, "0")},
			expectedDto: ResultDto{Text: "a.go", Turns: 2, SessionID: "session-3", CostUSD: "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockQueryProxy)
			proxy.On("Ask", mock.Anything, tt.expectedQuery).Return(tt.outcome, nil)

			dto, err := NewUseCase(proxy).Ask(context.Background(), tt.request)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedDto, dto)
			proxy.AssertExpectations(t)
		})
	}
}

func TestUseCase_Ask_Refusals(t *testing.T) {
	tests := []struct {
		name          string
		request       Request
		expectedError error
		expectedText  string
	}{
		{name: "an empty request", request: Request{Prompt: "  \n\t "}, expectedError: ErrEmptyRequest, expectedText: "the request is empty"},
		{name: "a negative round limit", request: Request{Prompt: "hi", MaxTurns: -1}, expectedError: ErrNegativeMaxTurns, expectedText: "the round limit cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockQueryProxy)

			dto, err := NewUseCase(proxy).Ask(context.Background(), tt.request)

			require.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedText, err.Error())
			assert.Equal(t, ResultDto{}, dto)
			proxy.AssertNotCalled(t, "Ask", mock.Anything, mock.Anything)
		})
	}
}

func TestUseCase_Ask_Failures(t *testing.T) {
	tests := []struct {
		name          string
		outcome       OutcomeVo
		proxyError    error
		expectedError error
		expectedText  string
	}{
		{
			name:          "not signed in",
			outcome:       OutcomeVo{Kind: OutcomeNotSignedIn},
			expectedError: ErrNotSignedIn,
			expectedText:  "not signed in",
		},
		{
			name:          "an unavailable model is reported with claude's explanation",
			outcome:       OutcomeVo{Kind: OutcomeProblem, Explanation: "There's an issue with the selected model (no-such-model)."},
			expectedError: ErrClaudeReportedProblem,
			expectedText:  "claude reported a problem: There's an issue with the selected model (no-such-model).",
		},
		{
			name:          "a missing work folder",
			proxyError:    ErrWorkFolderMissing,
			expectedError: ErrWorkFolderMissing,
			expectedText:  "the work folder does not exist",
		},
		{
			name:          "an unknown conversation",
			proxyError:    ErrConversationNotFound,
			expectedError: ErrConversationNotFound,
			expectedText:  "conversation not found",
		},
		{
			name:          "giving up",
			proxyError:    fmt.Errorf("claude code was given up: %w", context.Canceled),
			expectedError: context.Canceled,
			expectedText:  "given up",
		},
		{
			name:          "an unreadable reply",
			proxyError:    &UnreadableReplyError{Output: "garbage"},
			expectedError: ErrUnreadableReply,
			expectedText:  `could not read the reply: claude code said "garbage"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockQueryProxy)
			proxy.On("Ask", mock.Anything, QueryVo{Prompt: "hi"}).Return(tt.outcome, tt.proxyError)

			dto, err := NewUseCase(proxy).Ask(context.Background(), Request{Prompt: "hi"})

			require.ErrorIs(t, err, tt.expectedError)
			assert.Contains(t, err.Error(), tt.expectedText)
			assert.Equal(t, ResultDto{}, dto)
		})
	}
}

func TestUseCase_Ask_AnOutcomeWithoutAKindIsAProblem(t *testing.T) {
	proxy := new(mockQueryProxy)
	proxy.On("Ask", mock.Anything, QueryVo{Prompt: "hi"}).Return(OutcomeVo{Text: "OK"}, nil)

	_, err := NewUseCase(proxy).Ask(context.Background(), Request{Prompt: "hi"})

	require.ErrorIs(t, err, ErrClaudeReportedProblem)
}

func TestUseCase_Ask_RoundLimitReachedStillReportsTheWork(t *testing.T) {
	proxy := new(mockQueryProxy)
	proxy.On("Ask", mock.Anything, QueryVo{Prompt: "Run ls", MaxTurns: 1}).Return(OutcomeVo{
		Kind: OutcomeRoundLimitReached, Turns: 2, SessionID: "session-4", Cost: cost(t, "0.0055135800000000006"),
	}, nil)

	_, err := NewUseCase(proxy).Ask(context.Background(), Request{Prompt: "Run ls", MaxTurns: 1})

	require.ErrorIs(t, err, ErrRoundLimitReached)
	var limit *RoundLimitError
	require.ErrorAs(t, err, &limit)
	assert.Equal(t, RoundLimitError{Turns: 2, CostUSD: "0.00551358", CostNanodollars: 5_513_580, SessionID: "session-4"}, *limit)
	assert.Equal(t, "round limit reached after 2 rounds", err.Error())
}

func TestUseCase_Ask_DoesNotShareTheCallersToolList(t *testing.T) {
	tools := []string{"Read"}
	proxy := new(mockQueryProxy)
	proxy.On("Ask", mock.Anything, mock.Anything).Return(OutcomeVo{Kind: OutcomeFinished}, nil).Run(func(args mock.Arguments) {
		tools[0] = "Bash"
		assert.Equal(t, []string{"Read"}, args.Get(1).(QueryVo).AllowedTools)
	})

	_, err := NewUseCase(proxy).Ask(context.Background(), Request{Prompt: "hi", AllowedTools: tools})

	require.NoError(t, err)
}
