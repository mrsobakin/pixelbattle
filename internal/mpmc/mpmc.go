package mpmc

import (
	"fmt"
	"sync"
)

type MPMC[T any] struct {
	lock     sync.Mutex
	notify   *sync.Cond
	idx      uint64
	capacity uint64
	buffer   []T
}

func NewMPMC[T any](capacity uint64) *MPMC[T] {
	return &MPMC[T]{
		notify:   sync.NewCond(&sync.Mutex{}),
		idx:      0,
		capacity: capacity,
		buffer:   make([]T, capacity),
	}
}

func (mpmc *MPMC[T]) Send(value T) {
	mpmc.lock.Lock()
	mpmc.notify.L.Lock()
	mpmc.buffer[mpmc.idx%mpmc.capacity] = value
	mpmc.idx += 1
	mpmc.notify.Broadcast()
	mpmc.notify.L.Unlock()
	mpmc.lock.Unlock()
}

type Consumer[T any] struct {
	mpmc *MPMC[T]
	last uint64
}

func (mpmc *MPMC[T]) Subscribe() *Consumer[T] {
	return &Consumer[T]{
		mpmc: mpmc,
		last: mpmc.idx,
	}
}

func (c *Consumer[T]) Receive() (*T, error) {
	c.mpmc.lock.Lock()

	if (c.mpmc.idx - c.last) > c.mpmc.capacity {
		c.last = c.mpmc.idx
		c.mpmc.lock.Unlock()
		return nil, fmt.Errorf("lagging")
	}

	if c.mpmc.idx == c.last {
		c.mpmc.lock.Unlock()
		c.mpmc.notify.L.Lock()
		c.mpmc.notify.Wait()
		c.mpmc.notify.L.Unlock()
		return c.Receive()
	}

	idx := c.last % c.mpmc.capacity
	c.last += 1

	defer c.mpmc.lock.Unlock()
	return &c.mpmc.buffer[idx], nil
}
