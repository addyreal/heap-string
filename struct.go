package heapstr

import (
	"runtime"
)

type Buffer struct {
	b []byte
}

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
}
