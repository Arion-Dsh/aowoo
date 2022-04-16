//go:build darwin
// +build darwin

package aowoo

import (
	"sync"

	"github.com/arion-dsh/aowoo/avaudio"
)

func newDriver(rate, depth, chs int) driver {
	d := &driverTmp{
		cond:      sync.NewCond(new(sync.Mutex)),
		buff:      []float32{},
		rate:      rate,
		depth:     depth,
		chs:       chs,
		framesNum: 4096,
		device:    avaudio.New(),
	}
	return d
}
