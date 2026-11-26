package query

import (
	"fmt"
	"strconv"
	"strings"
)

const nanodollarsPerDollar = 1_000_000_000

// CostVo is an amount of US dollars, exact to the billionth of a dollar.
type CostVo struct {
	nanodollars int64
}

// NewCostVo reads a non-negative decimal amount of dollars, such as
// "0.0055135800000000006". Digits beyond the billionth are dropped, not rounded.
func NewCostVo(text string) (CostVo, error) {
	whole, fraction, _ := strings.Cut(strings.TrimSpace(text), ".")
	notDigit := func(r rune) bool { return r < '0' || r > '9' }
	if whole == "" || strings.IndexFunc(whole+fraction, notDigit) >= 0 {
		return CostVo{}, fmt.Errorf("%w: %q", ErrInvalidCost, text)
	}
	fraction = (fraction + "000000000")[:9]
	dollars, wholeErr := strconv.ParseInt(whole, 10, 64)
	billionths, fractionErr := strconv.ParseInt(fraction, 10, 64)
	if wholeErr != nil || fractionErr != nil || dollars > (1<<63-1)/nanodollarsPerDollar {
		return CostVo{}, fmt.Errorf("%w: %q", ErrInvalidCost, text)
	}
	return CostVo{nanodollars: dollars*nanodollarsPerDollar + billionths}, nil
}

// Nanodollars is the amount in billionths of a dollar.
func (c CostVo) Nanodollars() int64 {
	return c.nanodollars
}

// String writes the amount in dollars without trailing zeros, such as "0.00551358" or "2".
func (c CostVo) String() string {
	whole := c.nanodollars / nanodollarsPerDollar
	fraction := strings.TrimRight(fmt.Sprintf("%09d", c.nanodollars%nanodollarsPerDollar), "0")
	if fraction == "" {
		return strconv.FormatInt(whole, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + fraction
}

// QueryVo is a validated request handed to the infrastructure.
type QueryVo struct {
	Prompt           string
	MaxTurns         int
	Model            string
	WorkingDirectory string
	AllowedTools     []string
	SessionID        string
}

// OutcomeKind is how Claude's work on a request ended.
type OutcomeKind int

const (
	// OutcomeFinished means Claude gave its final reply.
	OutcomeFinished OutcomeKind = iota
	// OutcomeRoundLimitReached means Claude stopped at the round limit.
	OutcomeRoundLimitReached
	// OutcomeNotSignedIn means nobody is signed in.
	OutcomeNotSignedIn
	// OutcomeProblem means Claude reported a problem with an explanation.
	OutcomeProblem
)

// OutcomeVo is Claude Code's final answer as normalized by the infrastructure.
type OutcomeVo struct {
	Kind        OutcomeKind
	Text        string
	Turns       int
	SessionID   string
	Cost        CostVo
	Explanation string
}

// ToDomain converts the answer into the domain model that decides what it means.
func (vo OutcomeVo) ToDomain() OutcomeDomain {
	return OutcomeDomain{outcome: vo}
}
