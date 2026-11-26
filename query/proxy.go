package query

import "context"

// iQueryProxy runs one request through Claude Code and reports how it ended.
//
// Ask returns ErrWorkFolderMissing before running when the work folder does not
// exist, ErrConversationNotFound for an unknown conversation, and an error
// matching the context's error when the context ends first.
//
// Open starts a conversation; it returns ErrWorkFolderMissing before starting
// when the work folder does not exist.
type iQueryProxy interface {
	Ask(ctx context.Context, query QueryVo) (OutcomeVo, error)
	Open(ctx context.Context, query QueryVo) (IConversationLink, error)
}

// IConversationLink is one open conversation with Claude Code. The library
// implements it; programs never need to.
//
// Send writes one message and waits for its final answer. When ctx ends first
// it interrupts the reply and returns an error matching the context's error.
// When Claude Code has stopped, Send returns ErrConversationEnded, or
// ErrConversationNotFound when Claude Code did not know the conversation.
// Interrupt asks Claude Code to abandon the reply being written. Close ends
// Claude Code's work and may be called more than once.
type IConversationLink interface {
	Send(ctx context.Context, message string) (OutcomeVo, error)
	Interrupt(ctx context.Context) error
	Close() error
}
