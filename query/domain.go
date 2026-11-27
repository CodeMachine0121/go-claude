package query

import "strings"

// QueryDomain is a request that has been checked before anything is sent.
type QueryDomain struct {
	query QueryVo
}

// NewQueryDomain refuses empty requests and negative round limits.
func NewQueryDomain(request Request) (QueryDomain, error) {
	if strings.TrimSpace(request.Prompt) == "" {
		return QueryDomain{}, ErrEmptyRequest
	}
	if request.MaxTurns < 0 {
		return QueryDomain{}, ErrNegativeMaxTurns
	}
	return QueryDomain{query: QueryVo{
		Prompt:                 request.Prompt,
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
func (d QueryDomain) ToVo() QueryVo {
	return d.query
}

// OutcomeDomain decides what Claude's final answer means for the developer.
type OutcomeDomain struct {
	outcome OutcomeVo
}

// ToDto returns the final reply, or the failure the outcome stands for. A
// reached round limit or usage limit still reports rounds, cost and conversation.
func (d OutcomeDomain) ToDto() (ResultDto, error) {
	switch d.outcome.Kind {
	case OutcomeFinished:
		return ResultDto{
			Text:            d.outcome.Text,
			Turns:           d.outcome.Turns,
			SessionID:       d.outcome.SessionID,
			CostUSD:         d.outcome.Cost.String(),
			CostNanodollars: d.outcome.Cost.Nanodollars(),
		}, nil
	case OutcomeRoundLimitReached:
		return ResultDto{}, &RoundLimitError{
			Turns:           d.outcome.Turns,
			CostUSD:         d.outcome.Cost.String(),
			CostNanodollars: d.outcome.Cost.Nanodollars(),
			SessionID:       d.outcome.SessionID,
		}
	case OutcomeUsageLimitReached:
		return ResultDto{}, &UsageLimitError{
			ResetsAt:        d.outcome.ResetsAt,
			Explanation:     d.outcome.Explanation,
			Turns:           d.outcome.Turns,
			CostUSD:         d.outcome.Cost.String(),
			CostNanodollars: d.outcome.Cost.Nanodollars(),
			SessionID:       d.outcome.SessionID,
		}
	case OutcomeNotSignedIn:
		return ResultDto{}, ErrNotSignedIn
	case OutcomeInterrupted:
		return ResultDto{}, ErrInterrupted
	case OutcomeConversationNotFound:
		return ResultDto{}, ErrConversationNotFound
	default:
		return ResultDto{}, &ProblemError{Explanation: d.outcome.Explanation}
	}
}
