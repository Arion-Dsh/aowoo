package aowoo

import (
	"sync"
)

type driver interface {
	setCallback(callback)
	paused(state)
}

type callback func(b []float32) (int, error)

type device interface {
	Open(rate, depth, chs, framesNum int, read func([]float32, int)) error
	Pause(state int)
}

type driverTmp struct {
	cond *sync.Cond
	buff []float32

	rate      int
	depth     int
	chs       int
	framesNum int

	ready  bool
	device device
}

func (d *driverTmp) paused(s state) {
	d.device.Pause(int(s))
}

func (d *driverTmp) setCallback(f callback) {

	go func() {
		d.device.Open(d.rate, d.depth, d.chs, d.framesNum, d.read)

		for {
			d.cond.L.Lock()
			data := make([]float32, d.framesNum*d.chs*4)
			n, _ := f(data)
			if n == 0 {

				d.cond.Signal()
				d.cond.L.Unlock()
				continue
			}
			d.buff = append(d.buff, data...)
			d.cond.Broadcast()
			d.cond.L.Unlock()
		}
	}()
}

func (o *driverTmp) read(buf []float32, l int) {

	if len(o.buff) < l {
		l = len(o.buff)
	}

	for i := 0; i < l; i++ {
		buf[i] = o.buff[i]
	}
	o.buff = o.buff[l:]
}
