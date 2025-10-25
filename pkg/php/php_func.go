package php

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// 后续还需其它拓展,可在这个基础上实现

// BcAdd adds two Decimals.
func BcAdd(d1, d2 decimal.Decimal, precision int32) decimal.Decimal {
	if precision > 0 {
		return d1.Add(d2).RoundUp(precision)
	}
	return d1.Add(d2)
}

// BcSub subtracts two Decimals.
func BcSub(d1, d2 decimal.Decimal, precision int32) decimal.Decimal {
	if precision > 0 {
		return d1.Sub(d2).RoundUp(precision)
	}
	return d1.Sub(d2)
}

// BcMul multiplies two Decimals.
func BcMul(d1, d2 decimal.Decimal, precision int32) decimal.Decimal {
	if precision > 0 {
		return d1.Mul(d2).RoundUp(precision)
	}
	return d1.Mul(d2)
}

// BcDiv divides two Decimals.
func BcDiv(d1, d2 decimal.Decimal, precision int32) (decimal.Decimal, error) {
	if d2.Sign() == 0 {
		return decimal.Zero, fmt.Errorf("division by zero")
	}

	if precision > 0 {
		return d1.Div(d2).RoundUp(precision), nil
	}
	return d1.Div(d2), nil
}

// BcCmp compares two Decimals.
func BcCmp(d1, d2 decimal.Decimal) decimal.Decimal {
	if d1.GreaterThanOrEqual(d2) {
		return d1
	}
	return d2
}

// BcEqual reports whether d1 and d2 are equal.
func BcEqual(d1, d2 decimal.Decimal) bool {
	return d1.Equal(d2)
}

// BcZero decimal.Zero
func BcZero() decimal.Decimal {
	return decimal.Zero
}
