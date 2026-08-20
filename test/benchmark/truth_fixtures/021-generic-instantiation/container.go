// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package container

type Container[T any] struct {
	items []T
}

func (c *Container[T]) Push(item T) {
	c.items = append(c.items, item)
}

func (c *Container[T]) Pop() (T, bool) {
	var zero T
	if len(c.items) == 0 {
		return zero, false
	}
	item := c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]
	return item, true
}

func ProcessWorkflow() string {
	c := &Container[string]{}
	c.Push("garuda-semantic")
	val, _ := c.Pop()
	return val
}
