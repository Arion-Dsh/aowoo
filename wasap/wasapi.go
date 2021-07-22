//go:build windows
// +build windows

package wasap

/*
#cgo LDFLAGS: -lole32 -lwinmm -lksuser -luuid

#include <windows.h>
#include <mmsystem.h>

int aowoo_open(HWAVEOUT **waveOut, int rate, int depth, int chs, void *ctx);
void read_data(float *buuf, int l, void *ctx_);
*/
import "C"
import (
	"reflect"
	"sync"
	"unsafe"
)

var theDevice *device

func New() *device {
	if theDevice == nil {
		theDevice = &device{
			ctx: &ctx{1},
		}
	}
	return theDevice
}

type ctx struct {
	id int
}

type device struct {
	waveout *C.HWAVEOUT
	ctx     *ctx
	isPause bool

	mu       sync.Mutex
	readFunc func([]float32, int)
}

func (d *device) Open(rate, depth, chs, numFrames int, fc func([]float32, int)) error {

	//#TODO  open more device
	d.readFunc = fc
	go func() {
		C.aowoo_open(
			&d.waveout,
			C.int(rate),
			C.int(depth),
			C.int(chs),
			unsafe.Pointer(d.ctx),
		)
	}()

	return nil

}

func (d *device) Pause(state int) {

	d.mu.Lock()
	defer d.mu.Unlock()
	if state == 1 && !theDevice.isPause {
		// C.aowoo_pause(theDevice.queue, C.char(1))
	}

	if state != 1 && theDevice.isPause {
		// C.aowoo_pause(theDevice.queue, C.char(0))
	}

}

//export read_data
func read_data(buff *C.float, l C.int, ctx_ unsafe.Pointer) {

	buf := make([]float32, int(l))
	h := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	h.Data = uintptr(unsafe.Pointer(buff))
	h.Len = int(l)
	h.Cap = int(l)
	for i := range buf {
		buf[i] = 0
	}
	theDevice.readFunc(buf, int(l))
}
