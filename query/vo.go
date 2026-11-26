package query

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const nanodollarsPerDollar = 1_000_000_000

// CostVo is an amount of US dollars, exact to the billionth of a dollar.
type CostVo struct {
	nanodollars int64
}

// maximumExponent bounds exponent notation: larger amounts are not real costs,
// and the bound keeps a hostile exponent from allocating a huge digit string.
// Smaller amounts are far below a billionth of a dollar and read as zero.
const maximumExponent = 30

// NewCostVo reads a non-negative decimal amount of dollars, such as
// "0.0055135800000000006" or, as Claude Code writes very small amounts, "1.5e-7".
// Digits beyond the billionth are dropped, not rounded. The amount is never
// handled as a floating-point number.
func NewCostVo(text string) (CostVo, error) {
	invalid := fmt.Errorf("%w: %q", ErrInvalidCost, text)
	mantissa, exponentText, hasExponent := strings.Cut(strings.ToLower(strings.TrimSpace(text)), "e")
	exponent := 0
	if hasExponent {
		parsed, err := strconv.Atoi(exponentText)
		if err != nil || parsed > maximumExponent {
			return CostVo{}, invalid
		}
		exponent = max(parsed, -maximumExponent)
	}
	whole, fraction, _ := strings.Cut(mantissa, ".")
	notDigit := func(r rune) bool { return r < '0' || r > '9' }
	if whole == "" || strings.IndexFunc(whole+fraction, notDigit) >= 0 {
		return CostVo{}, invalid
	}
	digits := whole + fraction
	point := len(whole) + exponent
	if point <= 0 {
		digits, point = strings.Repeat("0", 1-point)+digits, 1
	}
	if point > len(digits) {
		digits += strings.Repeat("0", point-len(digits))
	}
	dollarsText := strings.TrimLeft(digits[:point], "0")
	billionthsText := (digits[point:] + "000000000")[:9]
	dollars := int64(0)
	if dollarsText != "" {
		parsed, err := strconv.ParseInt(dollarsText, 10, 64)
		if err != nil {
			return CostVo{}, invalid
		}
		dollars = parsed
	}
	// The billionths are exactly nine digits, so they always parse.
	billionths, _ := strconv.ParseInt(billionthsText, 10, 64)
	if dollars > (math.MaxInt64-billionths)/nanodollarsPerDollar {
		return CostVo{}, invalid
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

// Since is the increase of a running total since previous. A total that went
// down means counting restarted, so the whole total is the increase.
func (c CostVo) Since(previous CostVo) CostVo {
	if c.nanodollars < previous.nanodollars {
		return c
	}
	return CostVo{nanodollars: c.nanodollars - previous.nanodollars}
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
	// OutcomeFinished means Claude gave its final reply. The zero OutcomeKind
	// is none of the kinds below and is treated as a problem.
	OutcomeFinished OutcomeKind = iota + 1
	// OutcomeRoundLimitReached means Claude stopped at the round limit.
	OutcomeRoundLimitReached
	// OutcomeNotSignedIn means nobody is signed in.
	OutcomeNotSignedIn
	// OutcomeProblem means Claude reported a problem with an explanation.
	OutcomeProblem
	// OutcomeInterrupted means the reply was abandoned by an interrupt.
	OutcomeInterrupted
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
