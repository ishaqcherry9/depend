package php

import (
	testing2 "testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBcAdd(t *testing2.T) {
	d1 := decimal.NewFromFloat(1.234)
	d2 := decimal.NewFromFloat(5.678)

	result := BcAdd(d1, d2, 0)
	assert.Equal(t, "6.912", result.String())

	result = BcAdd(d1, d2, 2)
	assert.Equal(t, "6.92", result.String())

	result = BcAdd(d1, d2, -1)
	assert.Equal(t, "6.912", result.String()) // Should behave like precision 0
}

func TestBcSub(t *testing2.T) {
	d1 := decimal.NewFromFloat(5.678)
	d2 := decimal.NewFromFloat(1.234)

	result := BcSub(d1, d2, 0)
	assert.Equal(t, "4.444", result.String())

	result = BcSub(d1, d2, 2)
	assert.Equal(t, "4.45", result.String())

	result = BcSub(d1, d2, -1)
	assert.Equal(t, "4.444", result.String()) // Should behave like precision 0
}

func TestBcMul(t *testing2.T) {
	d1 := decimal.NewFromFloat(1.5)
	d2 := decimal.NewFromFloat(2.5)

	result := BcMul(d1, d2, 0)
	assert.Equal(t, "3.75", result.String())

	result = BcMul(d1, d2, 1)
	assert.Equal(t, "3.8", result.String())

	result = BcMul(d1, d2, -1)
	assert.Equal(t, "3.75", result.String()) // Should behave like precision 0
}

func TestBcDiv(t *testing2.T) {
	d1 := decimal.NewFromFloat(7.5)
	d2 := decimal.NewFromFloat(2.5)

	result, err := BcDiv(d1, d2, 0)
	assert.NoError(t, err)
	assert.Equal(t, "3", result.String())

	result, err = BcDiv(d1, d2, 1)
	assert.NoError(t, err)
	assert.Equal(t, "3", result.String())

	result, err = BcDiv(d1, d2, -1)
	assert.NoError(t, err)
	assert.Equal(t, "3", result.String()) // Should behave like precision 0

	// Division by zero
	d3 := decimal.NewFromInt(0)
	_, err = BcDiv(d1, d3, 2)
	assert.Error(t, err)
	assert.Equal(t, "division by zero", err.Error())
}

func TestBcCmp(t *testing2.T) {
	d1 := decimal.NewFromFloat(5)
	d2 := decimal.NewFromFloat(2.5)
	d3 := decimal.NewFromFloat(5)

	result := BcCmp(d1, d2)
	assert.Equal(t, "5", result.String())

	result = BcCmp(d2, d1)
	assert.Equal(t, "5", result.String())

	result = BcCmp(d1, d3)
	assert.Equal(t, "5", result.String())
}

func TestBcEqual(t *testing2.T) {
	d1 := decimal.NewFromFloat(5)
	d2 := decimal.NewFromFloat(2.5)
	d3 := decimal.NewFromFloat(5)

	assert.False(t, BcEqual(d1, d2))
	assert.True(t, BcEqual(d1, d3))
}

func TestBcZero(t *testing2.T) {
	zero := BcZero()
	assert.Equal(t, "0", zero.String())
}
