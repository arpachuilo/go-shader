package main

import "errors"

// CyclicArray
type CyclicArray[T any] struct {
	index int
	data  []T
}

func NewCyclicArray[T any](data []T) *CyclicArray[T] {
	index := 0
	if len(data) == 0 {
		index = -1
	}

	return &CyclicArray[T]{
		index: index,
		data:  data,
	}
}

// OffsetIndex set current index (wraps) and return it
func (self *CyclicArray[T]) OffsetIndex(offset int) int {
	self.index += offset
	if self.index > len(self.data)-1 {
		self.index = 0
	}

	if self.index < 0 {
		self.index = len(self.data) - 1
	}

	return self.index
}

func (self *CyclicArray[T]) Add(v T) {
	self.data = append(self.data, v)
}

func (self *CyclicArray[T]) Remove(i int) error {
	if i < 0 || i >= len(self.data) {
		return errors.New("out of bounds index")
	}

	self.data = append(self.data[:i], self.data[i+1:]...)
	self.OffsetIndex(0)
	return nil
}

func (self CyclicArray[T]) Index() int {
	return self.index
}

func (self CyclicArray[T]) Current() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.index]
}

func (self *CyclicArray[T]) Next() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.OffsetIndex(1)]
}

func (self *CyclicArray[T]) Previous() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.OffsetIndex(-1)]
}
