package main

import "errors"

// CyclableInt cyclable int [min, max)
type CyclableInt struct {
	min   int
	max   int
	index int
}

// NewCyclableInt create new cyclable
func NewCyclableInt(min, max int) *CyclableInt {
	return &CyclableInt{
		min:   min,
		max:   max,
		index: min,
	}
}

// Index return current index
func (c CyclableInt) Index() int {
	return c.index
}

// Min return current min
func (c CyclableInt) Min() int {
	return c.min
}

// Max return current max
func (c CyclableInt) Max() int {
	return c.max
}

// SetIndex set current index (wraps) and return it
func (c *CyclableInt) SetIndex(i int) int {
	c.index = i
	if c.index > c.max-1 {
		c.index = c.min
	}

	if c.index < c.min {
		c.index = c.max - 1
	}

	return c.index
}

func (c *CyclableInt) SetRange(min, max int) error {
	if min <= max {
		c.max = max
		c.min = min
		return nil
	}

	return errors.New("cannot set min greater than max")
}

func (c *CyclableInt) SetMin(min int) error {
	if min > c.max {
		return errors.New("cannot set min to value greater than min")
	}

	c.max = min
	return nil
}

func (c *CyclableInt) SetMax(max int) error {
	if max <= c.min {
		return errors.New("cannot set max to value smaller or equal to min")
	}

	c.max = max
	return nil
}

// Next go to next available index and return it
func (c *CyclableInt) Next() int {
	return c.SetIndex(c.index + 1)
}

// Prev go to previous available index and return it
func (c *CyclableInt) Prev() int {
	return c.SetIndex(c.index - 1)
}
