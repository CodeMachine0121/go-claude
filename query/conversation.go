package query

import (
	"context"
	"errors"
	"sync"
)

// ConversationRequest opens a conversation. Every message in it uses these choices.
type ConversationRequest struct {
	// MaxTurns caps the rounds Claude may take for each message. 0 means no cap.
	MaxTurns int
	// Model chooses the Claude model; empty means Claude Code's default.
	Model string
	// WorkingDirectory is the project folder Claude works in; it must exist.
	WorkingDirectory string
	// AllowedTools lists the tools Claude may use without asking.
	AllowedTools []string
	// SessionID continues an earlier conversation; empty starts a new one.
	SessionID string
	// MCPConfigurationPaths lists files describing the MCP servers Claude may
	// connect to; see Request.MCPConfigurationPaths.
	MCPConfigurationPaths []string
	// StrictMCPConfiguration makes Claude use only the MCP servers of
	// MCPConfigurationPaths.
	StrictMCPConfiguration bool
	// DisallowedTools lists the tools Claude must never use.
	DisallowedTools []string
	// AppendSystemPrompt is added after Claude Code's own system prompt. A
	// continued conversation keeps the text it began with.
	AppendSystemPrompt string
}

// ConversationDomain is a conversation request checked before anything starts.
type ConversationDomain struct {
	query QueryVo
}

// NewConversationDomain refuses negative round limits.
func NewConversationDomain(request ConversationRequest) (ConversationDomain, error) {
	if request.MaxTurns < 0 {
		return ConversationDomain{}, ErrNegativeMaxTurns
	}
	return ConversationDomain{query: QueryVo{
		MaxTurns:               request.MaxTurns,
		Model:                  request.Model,
		WorkingDirectory:       request.WorkingDirectory,
		AllowedTools:           append([]string(nil), request.AllowedTools...),
		SessionID:              request.SessionID,
		MCPConfigurationPaths:  append([]string(nil), request.MCPConfigurationPaths...),
		StrictMCPConfiguration: request.StrictMCPConfiguration,
		DisallowedTools:        append([]string(nil), request.DisallowedTools...),
		AppendSystemPrompt:     request.AppendSystemPrompt,
	}}, nil
}

// ToVo converts the checked request into the shape handed to the infrastructure.
func (d ConversationDomain) ToVo() QueryVo {
	return d.query
}

type conversationState int

const (
	conversationIdle conversationState = iota
	conversationAnswering
	conversationClosed
	conversationEnded
)

// refusal is why a message cannot be sent in this state; nil when it can.
func (s conversationState) refusal() error {
	switch s {
	case conversationAnswering:
		return ErrStillAnswering
	case conversationClosed:
		return ErrConversationClosed
	case conversationEnded:
		return ErrConversationEnded
	default:
		return nil
	}
}

// Conversation is an open conversation with Claude. One message is answered at
// a time. It is safe for concurrent use: Interrupt and Close may be called from
// another goroutine while Send waits. Always Close a conversation.
type Conversation struct {
	link      IConversationLink
	mutex     sync.Mutex
	state     conversationState
	total     CostVo
	sessionID string
}

// Send sends one message and returns Claude's final reply. ResultDto.CostUSD is
// what this reply cost; TotalCostUSD is the conversation's running total.
//
// It is refused with ErrEmptyRequest, ErrStillAnswering, ErrConversationClosed
// or ErrConversationEnded; it fails like Ask does otherwise, and with
// ErrInterrupted when Interrupt abandoned the reply. A *UsageLimitError leaves
// the conversation open: send again once the limit resets.
func (c *Conversation) Send(ctx context.Context, message string) (ResultDto, error) {
	c.mutex.Lock()
	if refusal := c.state.refusal(); refusal != nil {
		c.mutex.Unlock()
		return ResultDto{}, refusal
	}
	if _, err := NewQueryDomain(Request{Prompt: message}); err != nil {
		c.mutex.Unlock()
		return ResultDto{}, err
	}
	c.state = conversationAnswering
	c.mutex.Unlock()

	outcome, err := c.link.Send(ctx, message)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if outcome.Kind != 0 {
		// An answer arrived, even for a message given up on: it was spent.
		replyCost := outcome.Cost.Since(c.total)
		c.total = outcome.Cost
		if outcome.SessionID != "" {
			c.sessionID = outcome.SessionID
		}
		outcome.Cost = replyCost
	}
	if c.state == conversationClosed {
		return ResultDto{}, ErrConversationClosed
	}
	if errors.Is(err, ErrConversationEnded) || errors.Is(err, ErrConversationNotFound) {
		c.state = conversationEnded
		return ResultDto{}, err
	}
	c.state = conversationIdle
	if err != nil {
		return ResultDto{}, err
	}
	reply, err := outcome.ToDomain().ToDto()
	if errors.Is(err, ErrConversationNotFound) {
		c.state = conversationEnded
	}
	return reply, err
}

// Interrupt abandons the reply Claude is writing; the waiting Send fails with
// ErrInterrupted and the conversation stays open. When Claude is not answering,
// it does nothing. The next message cannot be sent before the interrupt is, so
// an interrupt never reaches a later message.
func (c *Conversation) Interrupt(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != conversationAnswering {
		return nil
	}
	return c.link.Interrupt(ctx)
}

// Close ends Claude Code's work for this conversation, including commands Claude
// started. Closing again does nothing.
func (c *Conversation) Close() error {
	c.mutex.Lock()
	if c.state == conversationClosed {
		c.mutex.Unlock()
		return nil
	}
	c.state = conversationClosed
	c.mutex.Unlock()
	return c.link.Close()
}

// SessionID identifies the conversation once Claude has replied; pass it in
// ConversationRequest.SessionID or Request.SessionID to continue it later.
func (c *Conversation) SessionID() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.sessionID
}

// TotalCostUSD is what the conversation has cost so far, as an exact decimal.
func (c *Conversation) TotalCostUSD() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.total.String()
}
