package github.com/arion-dsh/aowoo

import "sync/atomic"

var id uint32 = 0

func newAowooID() uint32 {
	return atomic.AddUint32(&id, 1)
}
