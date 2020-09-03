package common

import "sync"

type AtomicBool struct {
	mutex sync.Mutex
	value bool
}

func (b *AtomicBool) Value() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	v := b.value
	return v
}

func (b *AtomicBool) Set(v bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.value = v
}
