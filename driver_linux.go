//go:build linux
// +build linux

package aowoo

/*
#cgo LDFLAGS: -lasound

#include <alsa/asoundlib.h>
#include "alsa/alsa.c"

const char*
aowoo_open_device(snd_pcm_t **handle,
				int channels,
				int rate,
				snd_pcm_uframes_t *period_size);
*/
import "C"

import (
	"sync"
	"unsafe"
)

var (
	aowoo_handle *C.snd_pcm_t
	mu           sync.Mutex
)

func newDriver(rate, depth, chs int) driver {

	mu.Lock()
	defer mu.Unlock()

	a := &alsa{
		rate:  rate,
		depth: depth,
		chs:   chs,
	}

	var periodSize C.snd_pcm_uframes_t
	C.aowoo_open_device(&a.handle, C.int(chs), C.int(rate), &periodSize)
	a.periodSize = int(periodSize)

	return a
}

type alsa struct {
	handle     *C.snd_pcm_t
	rate       int
	chs        int
	depth      int
	periodSize int

	state chan int

	mu  sync.Mutex
	cnd *sync.Cond
}

func (a *alsa) paused(s state) {
	a.mu.Lock()
	defer a.mu.Unlock()
	C.snd_pcm_pause(a.handle, C.int(s))
}

func (a *alsa) close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	C.snd_pcm_close(a.handle)
}

func (a *alsa) setCallback(f callback) {

	go func() {
		for {
			buf := make([]float32, a.periodSize*a.chs)
			n, _ := f(buf)
			if n == 0 {
				continue
			}
			w := C.snd_pcm_writei(a.handle, unsafe.Pointer(&buf[0]), C.snd_pcm_uframes_t(a.periodSize))
			if w == -C.EPIPE {
				if errCode := C.snd_pcm_prepare(a.handle); errCode < 0 {
					panic("aowoo alsa writei errr")

				}
			}
		}
	}()

}
