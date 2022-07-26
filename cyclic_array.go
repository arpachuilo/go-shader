package main

import "errors"

// cyclicArray
type cyclicArray[T any] struct {
	index int
	data  []T
}

func newCyclicArray[T any](data []T) *cyclicArray[T] {
	index := 0
	if len(data) == 0 {
		index = -1
	}

	return &cyclicArray[T]{
		index: index,
		data:  data,
	}
}

// OffsetIndex set current index (wraps) and return it
func (self *cyclicArray[T]) OffsetIndex(offset int) int {
	self.index += offset
	if self.index > len(self.data)-1 {
		self.index = 0
	}

	if self.index < 0 {
		self.index = len(self.data) - 1
	}

	return self.index
}

func (self *cyclicArray[T]) Add(v T) {
	self.data = append(self.data, v)
}

func (self *cyclicArray[T]) Remove(i int) error {
	if i < 0 || i >= len(self.data) {
		return errors.New("out of bounds index")
	}

	self.data = append(self.data[:i], self.data[i+1:]...)
	self.OffsetIndex(0)
	return nil
}

func (self cyclicArray[T]) Index() int {
	return self.index
}

func (self cyclicArray[T]) Current() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.index]
}

func (self *cyclicArray[T]) Next() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.OffsetIndex(1)]
}

func (self *cyclicArray[T]) Previous() *T {
	if self.index == -1 {
		return nil
	}

	return &self.data[self.OffsetIndex(-1)]
}
