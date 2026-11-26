package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCostVo(t *testing.T) {
	tests := []struct {
		name                string
		text                string
		expectedText        string
		expectedNanodollars int64
	}{
		{name: "cost as claude code states it", text: "0.00559738", expectedText: "0.00559738", expectedNanodollars: 5_597_380},
		{name: "digits beyond the billionth are dropped", text: "0.0055135800000000006", expectedText: "0.00551358", expectedNanodollars: 5_513_580},
		{name: "dropped digits are not rounded up", text: "0.0000000019", expectedText: "0.000000001", expectedNanodollars: 1},
		{name: "nothing spent", text: "0", expectedText: "0", expectedNanodollars: 0},
		{name: "whole dollars", text: "2.50", expectedText: "2.5", expectedNanodollars: 2_500_000_000},
		{name: "whole dollars without a fraction", text: "3", expectedText: "3", expectedNanodollars: 3_000_000_000},
		{name: "very small amount in exponent form", text: "1e-7", expectedText: "0.0000001", expectedNanodollars: 100},
		{name: "exponent form with a fraction", text: "5.5E-7", expectedText: "0.00000055", expectedNanodollars: 550},
		{name: "exponent form below a billionth is nothing", text: "3e-10", expectedText: "0", expectedNanodollars: 0},
		{name: "far below a billionth is nothing", text: "1e-400", expectedText: "0", expectedNanodollars: 0},
		{name: "positive exponent past the digits", text: "1e3", expectedText: "1000", expectedNanodollars: 1_000_000_000_000},
		{name: "positive exponent", text: "1.25e2", expectedText: "125", expectedNanodollars: 125_000_000_000},
		{name: "largest amount that fits", text: "9223372036.854775807", expectedText: "9223372036.854775807", expectedNanodollars: 9_223_372_036_854_775_807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := NewCostVo(tt.text)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedText, cost.String())
			assert.Equal(t, tt.expectedNanodollars, cost.Nanodollars())
		})
	}
}

func TestNewCostVo_Invalid(t *testing.T) {
	for _, text := range []string{
		"", "-0.5", "+1", ".5", "0.12x", "0.123456789abc", "99999999999.0",
		"9223372036.9", "9223372036.854775808", "99999999999999999999", "1e400", "1e-x", "1e5e5", "-1e-7",
	} {
		t.Run(text, func(t *testing.T) {
			_, err := NewCostVo(text)

			require.ErrorIs(t, err, ErrInvalidCost)
		})
	}
}
