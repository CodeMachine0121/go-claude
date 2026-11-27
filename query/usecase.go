package query

import "context"

// UseCase sends one-shot requests. It is safe for concurrent use.
type UseCase struct {
	proxy iQueryProxy
}

// NewUseCase builds the use case on top of the infrastructure that runs Claude Code.
func NewUseCase(proxy iQueryProxy) *UseCase {
	return &UseCase{proxy: proxy}
}

// Ask sends the request and returns Claude's final reply. It blocks until Claude
// finishes, the round limit stops it (*RoundLimitError), the subscription's
// usage limit refuses it (*UsageLimitError), Claude reports a problem
// (*ProblemError, ErrNotSignedIn) or ctx ends. A missing or unreadable MCP
// configuration file fails with ErrInvalidMCPConfiguration.
func (u *UseCase) Ask(ctx context.Context, request Request) (ResultDto, error) {
	query, err := NewQueryDomain(request)
	if err != nil {
		return ResultDto{}, err
	}
	outcome, err := u.proxy.Ask(ctx, query.ToVo())
	if err != nil {
		return ResultDto{}, err
	}
	return outcome.ToDomain().ToDto()
}

// Open starts a conversation in which several messages can be sent. Close it
// when done. It is refused with ErrNegativeMaxTurns, ErrWorkFolderMissing or,
// for an MCP configuration file that does not exist, ErrInvalidMCPConfiguration.
func (u *UseCase) Open(ctx context.Context, request ConversationRequest) (*Conversation, error) {
	conversation, err := NewConversationDomain(request)
	if err != nil {
		return nil, err
	}
	link, err := u.proxy.Open(ctx, conversation.ToVo())
	if err != nil {
		return nil, err
	}
	return &Conversation{link: link, sessionID: request.SessionID}, nil
}
