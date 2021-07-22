package oboe

/*
#cgo CXXFLAGS: -std=c++17 -DOBOE_ENABLE_AAUDIO=0
#cgo LDFLAGS: -llog -lOpenSLES -static-libstdc++

#include "aowoo_oboe.h"
*/
import "C"
import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

var theDevice *device

func New() *device {
	if theDevice == nil {
		theDevice = &device{}
	}

	return theDevice
}

type device struct {
	mu       sync.Mutex
	isPause  bool
	readFunc func([]float32, int)
}

func (d *device) Open(rate, depth, chs, frames int, fc func(f []float32, l int)) error {
	d.readFunc = fc
	err := C.aowoo_open_oboe(C.int(rate), C.int(depth), C.int(chs), C.int(frames))
	return errformat(int(err))

}

func (d *device) Pause(state int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if state == 1 && d.isPause {
		return
	}
	C.aowoo_pause_oboe(C.int(state))
	if state == 1 {
		isPause = true
	}
}

func newErr(s string) error {
	return errors.New("aowoo oboe err: %s" + s)
}

func errformat(err int) error {
	switch err {
	case -900:
		return newErr("ErrorBase")
	case -886:
		return newErr("ErrorNull")
	case 882:
		return newErr("ErrorOutOfRange")
	}
	return nil
}

//export read_data
func read_data(buff *C.float, l C.int32_t) {
	var buf []float32

	h := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	h.Data = uintptr(unsafe.Pointer(buff))
	h.Len = int(l)
	h.Cap = int(l)
	for i := range buf {
		buf[i] = 0
	}
	theDevice.readFunc(buf, int(l))
}
