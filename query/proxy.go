package query

import "context"

// iQueryProxy runs one request through Claude Code and reports how it ended.
//
// Ask returns ErrWorkFolderMissing before running when the work folder does not
// exist, ErrConversationNotFound for an unknown conversation, and an error
// matching the context's error when the context ends first.
type iQueryProxy interface {
	Ask(ctx context.Context, query QueryVo) (OutcomeVo, error)
}
