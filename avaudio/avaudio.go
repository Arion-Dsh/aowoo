//go:build darwin
// +build darwin

package avaudio

/*
#cgo LDFLAGS: -framework AudioToolbox
#cgo LDFLAGS: -framework CoreAudio
#cgo LDFLAGS: -framework CoreServices

#include <AudioToolbox/AudioQueue.h>
#include <CoreAudio/CoreAudioTypes.h>
#include <CoreFoundation/CFRunLoop.h>


void read_data(void *buff, int l, void *id);
int aowoo_open(
		void *ctx,
        AudioQueueRef **queue,
        int rate,
        int chs);
void aowoo_pause(AudioQueueRef *q, char s);
void aowoo_close(AudioQueueRef *q);

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
	mu      sync.Mutex
	ctx     *ctx
	queue   *C.AudioQueueRef
	isPause bool

	readFunc func([]float32, int)
}

func (d *device) Open(rate, depth, chs, framesNum int, fc func([]float32, int)) error {
	d.readFunc = fc

	//#TODO  open more device
	go func() {
		C.aowoo_open(
			unsafe.Pointer(d.ctx),
			&d.queue,
			C.int(rate),
			C.int(chs))
	}()
	return nil
}

func (d *device) Pause(state int) {

	d.mu.Lock()
	defer d.mu.Unlock()
	if state == 1 && !d.isPause {
		C.aowoo_pause(d.queue, C.char(1))
	}

	if state != 1 && d.isPause {
		C.aowoo_pause(d.queue, C.char(0))
	}

}

//export read_data
func read_data(buff unsafe.Pointer, l C.int, ctx_ unsafe.Pointer) {

	var buf []float32
	h := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	h.Data = uintptr(buff)
	h.Len = int(l)
	h.Cap = int(l)
	for i := range buf {
		buf[i] = 0
	}
	theDevice.readFunc(buf, int(l))

}
